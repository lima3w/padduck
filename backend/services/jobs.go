package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"padduck/repository"
)

type JobStatus string

const (
	JobQueued    JobStatus = "queued"
	JobRunning   JobStatus = "running"
	JobSucceeded JobStatus = "succeeded"
	JobFailed    JobStatus = "failed"
	JobCanceled  JobStatus = "canceled"
)

type JobProgress struct {
	Current int64  `json:"current"`
	Total   int64  `json:"total"`
	Message string `json:"message,omitempty"`
}

type BackgroundJob struct {
	ID          int64                  `json:"id"`
	Kind        string                 `json:"kind"`
	Name        string                 `json:"name"`
	Status      JobStatus              `json:"status"`
	Progress    JobProgress            `json:"progress"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	Error       string                 `json:"error,omitempty"`
	Diagnostics []string               `json:"diagnostics,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Result      interface{}            `json:"result,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	FinishedAt  *time.Time             `json:"finished_at,omitempty"`

	cancel context.CancelFunc
	runner JobRunner
	runID  int64
}

type JobRunner func(context.Context, *JobReporter) (interface{}, error)

type JobReporter struct {
	service *JobService
	jobID   int64
}

func (r *JobReporter) Progress(current, total int64, message string) {
	r.service.setProgress(r.jobID, current, total, message)
}

func (r *JobReporter) Diagnostic(message string) {
	r.service.addDiagnostic(r.jobID, message)
}

type JobService struct {
	mu     sync.RWMutex
	nextID atomic.Int64
	jobs   map[int64]*BackgroundJob
	repo   *repository.Repository
}

// NewJobService creates a JobService. If repo is nil, jobs are tracked
// in-memory only (useful in tests). With a real repo, job state is persisted
// to the background_jobs table and stale running jobs are recovered on startup.
func NewJobService(repo *repository.Repository) *JobService {
	if repo != nil && !repo.HasPool() {
		repo = nil
	}
	s := &JobService{
		jobs: make(map[int64]*BackgroundJob),
		repo: repo,
	}
	if repo != nil {
		_ = repo.RecoverStaleJobs(context.Background())
	}
	return s
}

func (s *JobService) Enqueue(kind, name string, metadata map[string]interface{}, maxAttempts int, runner JobRunner) *BackgroundJob {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	now := time.Now().UTC()

	var id int64
	if s.repo != nil {
		var err error
		id, err = s.repo.InsertJob(context.Background(), kind, name, metadata, maxAttempts)
		if err != nil {
			id = s.nextID.Add(1)
		}
	} else {
		id = s.nextID.Add(1)
	}

	job := &BackgroundJob{
		ID:          id,
		Kind:        kind,
		Name:        name,
		Status:      JobQueued,
		MaxAttempts: maxAttempts,
		Metadata:    metadata,
		Diagnostics: []string{},
		CreatedAt:   now,
		runner:      runner,
	}

	s.mu.Lock()
	s.jobs[id] = job
	out := publicJob(job)
	s.mu.Unlock()

	go s.run(id)
	return out
}

// List returns all in-memory (active/recent) jobs plus historical DB records
// not currently in memory, sorted newest-first.
func (s *JobService) List() []*BackgroundJob {
	s.mu.RLock()
	inMemIDs := make(map[int64]struct{}, len(s.jobs))
	out := make([]*BackgroundJob, 0, len(s.jobs))
	for id, job := range s.jobs {
		inMemIDs[id] = struct{}{}
		out = append(out, publicJob(job))
	}
	s.mu.RUnlock()

	if s.repo != nil {
		records, err := s.repo.ListJobRecords(context.Background(), 100)
		if err == nil {
			for _, rec := range records {
				if _, ok := inMemIDs[rec.ID]; !ok {
					out = append(out, jobFromRecord(rec))
				}
			}
		}
	}

	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *JobService) Get(id int64) (*BackgroundJob, bool) {
	s.mu.RLock()
	job, ok := s.jobs[id]
	var out *BackgroundJob
	if ok {
		out = publicJob(job)
	}
	s.mu.RUnlock()
	if ok {
		return out, true
	}
	if s.repo != nil {
		rec, err := s.repo.GetJobRecord(context.Background(), id)
		if err == nil {
			return jobFromRecord(rec), true
		}
	}
	return nil, false
}

func (s *JobService) Cancel(id int64) (*BackgroundJob, error) {
	s.mu.Lock()
	job, ok := s.jobs[id]
	if !ok {
		s.mu.Unlock()
		// Check DB for historical job — can't cancel it, but report why.
		if s.repo != nil {
			rec, err := s.repo.GetJobRecord(context.Background(), id)
			if err == nil {
				if rec.Status != "queued" && rec.Status != "running" {
					return nil, errors.New("job is not cancellable")
				}
			}
		}
		return nil, fmt.Errorf("job not found")
	}
	if job.Status != JobQueued && job.Status != JobRunning {
		s.mu.Unlock()
		return nil, errors.New("job is not cancellable")
	}
	cancel := job.cancel
	job.Status = JobCanceled
	now := time.Now().UTC()
	job.FinishedAt = &now
	out := publicJob(job)
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if s.repo != nil {
		_ = s.repo.MarkJobCanceled(context.Background(), id)
	}
	return out, nil
}

func (s *JobService) Retry(id int64) (*BackgroundJob, error) {
	s.mu.Lock()
	job, ok := s.jobs[id]
	if !ok {
		s.mu.Unlock()
		return nil, fmt.Errorf("job not found or runner no longer available after restart")
	}
	if job.Status != JobFailed && job.Status != JobCanceled {
		s.mu.Unlock()
		return nil, errors.New("only failed or canceled jobs can be retried")
	}
	if job.runner == nil {
		s.mu.Unlock()
		return nil, errors.New("job cannot be retried")
	}
	job.Status = JobQueued
	job.Error = ""
	job.Result = nil
	job.Progress = JobProgress{}
	job.StartedAt = nil
	job.FinishedAt = nil
	job.runID++
	out := publicJob(job)
	s.mu.Unlock()
	go s.run(id)
	return out, nil
}

func (s *JobService) run(id int64) {
	s.mu.Lock()
	job, ok := s.jobs[id]
	if !ok || job.Status != JobQueued {
		s.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	now := time.Now().UTC()
	job.Status = JobRunning
	job.Attempts++
	job.runID++
	runID := job.runID
	job.StartedAt = &now
	job.FinishedAt = nil
	job.cancel = cancel
	runner := job.runner
	attempts := job.Attempts
	s.mu.Unlock()

	if s.repo != nil {
		_ = s.repo.MarkJobRunning(context.Background(), id, attempts)
	}

	result, err := runner(ctx, &JobReporter{service: s, jobID: id})

	s.mu.Lock()
	defer s.mu.Unlock()
	defer cancel()
	job = s.jobs[id]
	if job.runID != runID {
		return
	}

	var finalStatus JobStatus
	if job.Status == JobCanceled || errors.Is(err, context.Canceled) {
		finalStatus = JobCanceled
		job.Status = JobCanceled
	} else if err != nil {
		finalStatus = JobFailed
		job.Status = JobFailed
		job.Error = err.Error()
		job.Diagnostics = append(job.Diagnostics, err.Error())
	} else {
		finalStatus = JobSucceeded
		job.Status = JobSucceeded
		job.Result = result
		if job.Progress.Total > 0 {
			job.Progress.Current = job.Progress.Total
		}
	}
	finished := time.Now().UTC()
	job.FinishedAt = &finished
	job.cancel = nil

	if s.repo != nil {
		progress := 0
		if job.Progress.Total > 0 {
			progress = int(job.Progress.Current * 100 / job.Progress.Total)
		}
		diagnostic := strings.Join(job.Diagnostics, "\n")
		var resultBytes []byte
		if result != nil {
			resultBytes, _ = json.Marshal(result)
		}
		_ = s.repo.MarkJobFinished(context.Background(), id, string(finalStatus), job.Error, diagnostic, progress, resultBytes)
	}
}

func (s *JobService) setProgress(id, current, total int64, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if job, ok := s.jobs[id]; ok {
		job.Progress = JobProgress{Current: current, Total: total, Message: message}
	}
}

func (s *JobService) addDiagnostic(id int64, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if job, ok := s.jobs[id]; ok {
		job.Diagnostics = append(job.Diagnostics, message)
	}
}

func publicJob(job *BackgroundJob) *BackgroundJob {
	cp := *job
	cp.cancel = nil
	cp.runner = nil
	if job.Diagnostics != nil {
		cp.Diagnostics = append([]string(nil), job.Diagnostics...)
	}
	if job.Metadata != nil {
		cp.Metadata = make(map[string]interface{}, len(job.Metadata))
		for k, v := range job.Metadata {
			cp.Metadata[k] = v
		}
	}
	return &cp
}

// jobFromRecord converts a DB-sourced JobRecord into a BackgroundJob.
// The result has no runner or cancel func (the process that ran it may be gone).
func jobFromRecord(rec *repository.JobRecord) *BackgroundJob {
	job := &BackgroundJob{
		ID:          rec.ID,
		Kind:        rec.Kind,
		Name:        rec.Name,
		Status:      JobStatus(rec.Status),
		MaxAttempts: rec.MaxAttempts,
		Attempts:    rec.Attempts,
		Error:       rec.ErrorMessage,
		Metadata:    rec.Payload,
		Result:      rec.Result,
		CreatedAt:   rec.CreatedAt,
		StartedAt:   rec.StartedAt,
		FinishedAt:  rec.FinishedAt,
	}
	if rec.Progress > 0 {
		job.Progress = JobProgress{Current: int64(rec.Progress), Total: 100}
	}
	if rec.Diagnostic != "" {
		job.Diagnostics = strings.Split(rec.Diagnostic, "\n")
	} else {
		job.Diagnostics = []string{}
	}
	return job
}
