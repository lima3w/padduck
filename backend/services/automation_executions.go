package services

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"padduck/models"
	"padduck/repository"
)

// Errors returned by RollbackExecution.
var (
	ErrExecutionAlreadyRolledBack = errors.New("this execution has already been rolled back")
	ErrNothingToRollBack          = errors.New("this execution has no reversible actions")
)

// rollbackFunc reverses one successfully run action using the Undo data saved
// when it ran. It returns a status (rolled_back, skipped, not_reversible) and a
// message; an error means the rollback itself failed.
type rollbackFunc func(ctx context.Context, a *AutomationService, res models.PolicyActionResult) (status, message string, err error)

// rollbackRegistry holds the rollback hook for each action type. Actions that
// cannot be reversed are registered so the log says why, rather than silently
// doing nothing.
var rollbackRegistry = map[string]rollbackFunc{
	"tag":              rollbackTag,
	"scan":             rollbackScan,
	"webhook":          notReversible("a delivered webhook cannot be unsent"),
	"notify":           notReversible("a sent notification cannot be unsent"),
	"audit_annotation": notReversible("an audit annotation is a record and is kept"),
}

func notReversible(msg string) rollbackFunc {
	return func(context.Context, *AutomationService, models.PolicyActionResult) (string, string, error) {
		return "not_reversible", msg, nil
	}
}

// reversible reports whether rolling back this action type can change anything.
func reversible(actionType string) bool { return actionType == "tag" || actionType == "scan" }

// rollbackTag restores the tag the IP had before the action, but only while the
// action's tag is still the one applied; a later manual change is left alone.
func rollbackTag(ctx context.Context, a *AutomationService, res models.PolicyActionResult) (string, string, error) {
	ipID, err := strconv.ParseInt(res.Undo["ip_id"], 10, 64)
	if err != nil {
		return "", "", fmt.Errorf("rollback data is missing the IP address")
	}
	ip, err := a.repo.GetIPAddressByID(ctx, ipID)
	if err != nil {
		return "skipped", "the IP address no longer exists", nil
	}
	current := ""
	if ip.TagID != nil {
		current = strconv.FormatInt(*ip.TagID, 10)
	}
	if current != res.Undo["applied_tag_id"] {
		return "skipped", "the tag was changed after the policy applied it, so it was left as is", nil
	}
	var previous *int64
	if res.Undo["previous_tag_id"] != "" {
		v, err := strconv.ParseInt(res.Undo["previous_tag_id"], 10, 64)
		if err != nil {
			return "", "", fmt.Errorf("rollback data has an invalid previous tag")
		}
		previous = &v
	}
	if _, err := a.repo.UpdateIPAddressTag(ctx, ipID, previous); err != nil {
		return "", "", err
	}
	return "rolled_back", "restored the previous tag", nil
}

// rollbackScan removes the scan job the action created, as long as it has not run.
func rollbackScan(ctx context.Context, a *AutomationService, res models.PolicyActionResult) (string, string, error) {
	jobID, err := strconv.ParseInt(res.Undo["scan_job_id"], 10, 64)
	if err != nil {
		return "", "", fmt.Errorf("rollback data is missing the scan job")
	}
	job, err := a.repo.GetScanJobByID(ctx, jobID)
	if err != nil {
		return "skipped", "the scan job no longer exists", nil
	}
	if job.LastRunAt != nil {
		return "skipped", "the scan job has already run", nil
	}
	if err := a.repo.DeleteScanJob(ctx, jobID); err != nil {
		return "", "", err
	}
	return "rolled_back", "removed the queued scan job", nil
}

// ListExecutions returns execution log entries, newest first, plus the total.
func (a *AutomationService) ListExecutions(ctx context.Context, f repository.PolicyExecutionLogFilter) ([]*models.PolicyExecutionLog, int, error) {
	return a.repo.ListPolicyExecutionLogs(ctx, f)
}

// GetExecution returns one execution log entry.
func (a *AutomationService) GetExecution(ctx context.Context, id int64) (*models.PolicyExecutionLog, error) {
	return a.repo.GetPolicyExecutionLog(ctx, id)
}

// ExecutionCounts returns how many times each policy has matched.
func (a *AutomationService) ExecutionCounts(ctx context.Context) (map[int64]int, error) {
	return a.repo.CountPolicyExecutions(ctx)
}

// RollbackExecution reverses the reversible actions of one execution and
// records the outcome of each. A failure to reverse one action does not stop
// the others. It can only be done once per execution.
func (a *AutomationService) RollbackExecution(ctx context.Context, id int64, by *int64) (*models.PolicyExecutionLog, error) {
	l, err := a.repo.GetPolicyExecutionLog(ctx, id)
	if err != nil {
		return nil, err
	}
	if l.RolledBackAt != nil {
		return nil, ErrExecutionAlreadyRolledBack
	}
	reversibleFound := false
	for _, r := range l.ActionsTaken {
		if r.Status == "success" && reversible(r.Type) {
			reversibleFound = true
		}
	}
	if !reversibleFound {
		return nil, ErrNothingToRollBack
	}

	now := time.Now().UTC()
	actions := make([]models.PolicyActionResult, len(l.ActionsTaken))
	copy(actions, l.ActionsTaken)
	for i, r := range actions {
		if r.Status != "success" {
			continue
		}
		fn, ok := rollbackRegistry[r.Type]
		if !ok {
			fn = notReversible("no rollback is defined for this action type")
		}
		status, msg, err := fn(ctx, a, r)
		if err != nil {
			status, msg = "failed", err.Error()
		}
		actions[i].Rollback = &models.PolicyRollbackInfo{Status: status, Message: msg, At: now}
	}
	won, err := a.repo.MarkPolicyExecutionRolledBack(ctx, id, actions, by, now)
	if err != nil {
		return nil, err
	}
	if !won {
		return nil, ErrExecutionAlreadyRolledBack
	}
	return a.repo.GetPolicyExecutionLog(ctx, id)
}
