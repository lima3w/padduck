package services

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
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
}

func NewJobService() *JobService {
	return &JobService{jobs: make(map[int64]*BackgroundJob)}
}

func (s *JobService) Enqueue(kind, name string, metadata map[string]interface{}, maxAttempts int, runner JobRunner) *BackgroundJob {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	now := time.Now().UTC()
	job := &BackgroundJob{
		ID:          s.nextID.Add(1),
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
	s.jobs[job.ID] = job
	s.mu.Unlock()

	go s.run(job.ID)
	return publicJob(job)
}

func (s *JobService) List() []*BackgroundJob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*BackgroundJob, 0, len(s.jobs))
	for _, job := range s.jobs {
		out = append(out, publicJob(job))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *JobService) Get(id int64) (*BackgroundJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.jobs[id]
	if !ok {
		return nil, false
	}
	return publicJob(job), true
}

func (s *JobService) Cancel(id int64) (*BackgroundJob, error) {
	s.mu.Lock()
	job, ok := s.jobs[id]
	if !ok {
		s.mu.Unlock()
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
	return out, nil
}

func (s *JobService) Retry(id int64) (*BackgroundJob, error) {
	s.mu.Lock()
	job, ok := s.jobs[id]
	if !ok {
		s.mu.Unlock()
		return nil, fmt.Errorf("job not found")
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
	s.mu.Unlock()

	result, err := runner(ctx, &JobReporter{service: s, jobID: id})

	s.mu.Lock()
	defer s.mu.Unlock()
	defer cancel()
	job = s.jobs[id]
	if job.runID != runID {
		return
	}
	if job.Status == JobCanceled || errors.Is(err, context.Canceled) {
		job.Status = JobCanceled
	} else if err != nil {
		job.Status = JobFailed
		job.Error = err.Error()
		job.Diagnostics = append(job.Diagnostics, err.Error())
	} else {
		job.Status = JobSucceeded
		job.Result = result
		if job.Progress.Total > 0 {
			job.Progress.Current = job.Progress.Total
		}
	}
	finished := time.Now().UTC()
	job.FinishedAt = &finished
	job.cancel = nil
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
