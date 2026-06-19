package services

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestJobServiceTracksProgressAndResult(t *testing.T) {
	jobs := NewJobService(nil)
	job := jobs.Enqueue("test", "successful job", nil, 1, func(ctx context.Context, reporter *JobReporter) (interface{}, error) {
		reporter.Progress(1, 2, "halfway")
		reporter.Diagnostic("checkpoint")
		reporter.Progress(2, 2, "done")
		return map[string]string{"ok": "true"}, nil
	})

	got := waitForJob(t, jobs, job.ID)
	require.Equal(t, JobSucceeded, got.Status)
	require.Equal(t, int64(2), got.Progress.Current)
	require.Contains(t, got.Diagnostics, "checkpoint")
	require.NotNil(t, got.Result)
}

func TestJobServiceCancelAndRetry(t *testing.T) {
	jobs := NewJobService(nil)
	started := make(chan struct{})
	var once sync.Once
	job := jobs.Enqueue("test", "cancellable job", nil, 1, func(ctx context.Context, reporter *JobReporter) (interface{}, error) {
		once.Do(func() { close(started) })
		<-ctx.Done()
		return nil, ctx.Err()
	})
	<-started

	canceled, err := jobs.Cancel(job.ID)
	require.NoError(t, err)
	require.Equal(t, JobCanceled, canceled.Status)

	retried, err := jobs.Retry(job.ID)
	require.NoError(t, err)
	require.Equal(t, JobQueued, retried.Status)

	got := waitForStatus(t, jobs, job.ID, JobRunning)
	require.Equal(t, 2, got.Attempts)
	_, _ = jobs.Cancel(job.ID)
}

func TestJobServiceFailedJobCarriesDiagnostics(t *testing.T) {
	jobs := NewJobService(nil)
	job := jobs.Enqueue("test", "failed job", nil, 1, func(ctx context.Context, reporter *JobReporter) (interface{}, error) {
		return nil, errors.New("boom")
	})

	got := waitForJob(t, jobs, job.ID)
	require.Equal(t, JobFailed, got.Status)
	require.Equal(t, "boom", got.Error)
	require.Contains(t, got.Diagnostics, "boom")
}

func waitForJob(t *testing.T, jobs *JobService, id int64) *BackgroundJob {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		job, ok := jobs.Get(id)
		require.True(t, ok)
		if job.Status == JobSucceeded || job.Status == JobFailed || job.Status == JobCanceled {
			return job
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("job %d did not finish", id)
	return nil
}

func waitForStatus(t *testing.T, jobs *JobService, id int64, status JobStatus) *BackgroundJob {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		job, ok := jobs.Get(id)
		require.True(t, ok)
		if job.Status == status {
			return job
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("job %d did not reach status %s", id, status)
	return nil
}
