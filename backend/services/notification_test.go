package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"padduck/models"
)

type notificationRepoStub struct {
	user             *models.User
	prefs            *models.NotificationPreferences
	recentCount      int64
	pending          []*models.NotificationQueue
	queued           *models.NotificationQueue
	queuedData       map[string]interface{}
	getUserCalls     int
	getPrefsCalls    int
	countRecentCalls int
	createQueueCalls int
	markSent         []int64
	markFailed       []failedNotification
	cleanupCalls     int
	createQueueErr   error
	pendingErr       error
	markSentErr      error
	markFailedErr    error
	cleanupErr       error
}

type failedNotification struct {
	id          int64
	errMsg      string
	retryCount  int
	nextRetryAt *time.Time
}

func (r *notificationRepoStub) GetUserByID(context.Context, int64) (*models.User, error) {
	r.getUserCalls++
	return r.user, nil
}

func (r *notificationRepoStub) GetNotificationPreferences(context.Context, int64) (*models.NotificationPreferences, error) {
	r.getPrefsCalls++
	return r.prefs, nil
}

func (r *notificationRepoStub) CountRecentNotificationsSent(context.Context, int64, time.Time) (int64, error) {
	r.countRecentCalls++
	return r.recentCount, nil
}

func (r *notificationRepoStub) CreateNotificationQueueItem(_ context.Context, userID int64, email, template, dataJSON string) (*models.NotificationQueue, error) {
	r.createQueueCalls++
	if r.createQueueErr != nil {
		return nil, r.createQueueErr
	}
	_ = json.Unmarshal([]byte(dataJSON), &r.queuedData)
	r.queued = &models.NotificationQueue{ID: 99, UserID: userID, Email: email, Template: template, Data: dataJSON}
	return r.queued, nil
}

func (r *notificationRepoStub) GetPendingNotifications(context.Context, int) ([]*models.NotificationQueue, error) {
	return r.pending, r.pendingErr
}

func (r *notificationRepoStub) MarkNotificationSent(_ context.Context, id int64) error {
	r.markSent = append(r.markSent, id)
	return r.markSentErr
}

func (r *notificationRepoStub) MarkNotificationFailed(_ context.Context, id int64, errMsg string, retryCount int, nextRetryAt *time.Time) error {
	r.markFailed = append(r.markFailed, failedNotification{id: id, errMsg: errMsg, retryCount: retryCount, nextRetryAt: nextRetryAt})
	return r.markFailedErr
}

func (r *notificationRepoStub) CleanupOldNotifications(context.Context) error {
	r.cleanupCalls++
	return r.cleanupErr
}

type notificationEmailStub struct {
	configured bool
	appURL     string
	sendErr    error
	sent       []sentNotification
}

type sentNotification struct {
	to      string
	subject string
	body    string
}

func (e *notificationEmailStub) IsSMTPConfigured() bool { return e.configured }

func (e *notificationEmailStub) AppURL() string {
	if e.appURL == "" {
		return "http://localhost:3000"
	}
	return e.appURL
}

func (e *notificationEmailStub) Send(to, subject, body string) error {
	e.sent = append(e.sent, sentNotification{to: to, subject: subject, body: body})
	return e.sendErr
}

func TestRenderTemplate(t *testing.T) {
	svc := &NotificationService{}

	subject, body, err := svc.renderTemplate(NotifRequestRejected, map[string]interface{}{
		"Username":     "alice",
		"RequestType":  "subnet",
		"RequestID":    int64(42),
		"ReviewerNote": "CIDR is too broad",
	})

	require.NoError(t, err)
	assert.Equal(t, "Your subnet request has been rejected", subject)
	assert.Contains(t, body, "Hello alice")
	assert.Contains(t, body, "Request ID:    42")
	assert.Contains(t, body, "CIDR is too broad")
}

func TestRenderTemplateRejectsUnknownTemplate(t *testing.T) {
	_, _, err := (&NotificationService{}).renderTemplate("missing", nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown notification template")
}

func TestPreferenceEnabled(t *testing.T) {
	prefs := &models.NotificationPreferences{
		LoginSuccess:    false,
		LoginFailed:     true,
		AccountLocked:   true,
		PasswordChanged: true,
		MFAChanges:      false,
		APITokenChanges: true,
		RoleChanges:     false,
		SessionRevoked:  true,
	}

	assert.False(t, preferenceEnabled(prefs, NotifLoginSuccess))
	assert.True(t, preferenceEnabled(prefs, NotifLoginFailed))
	assert.False(t, preferenceEnabled(prefs, NotifMFAEnabled))
	assert.False(t, preferenceEnabled(prefs, NotifMFADisabled))
	assert.True(t, preferenceEnabled(prefs, NotifAPITokenCreated))
	assert.False(t, preferenceEnabled(prefs, NotifRoleChanged))
	assert.True(t, preferenceEnabled(prefs, "request-submitted"))
}

func TestRetryDelay(t *testing.T) {
	require.Equal(t, time.Minute, *retryDelay(0))
	require.Equal(t, 5*time.Minute, *retryDelay(1))
	require.Equal(t, 24*time.Hour, *retryDelay(5))
	assert.Nil(t, retryDelay(6))
}

func TestQueueSkipsWhenSMTPIsNotConfigured(t *testing.T) {
	repo := &notificationRepoStub{}
	email := &notificationEmailStub{configured: false}
	svc := &NotificationService{repo: repo, email: email}

	err := svc.Queue(context.Background(), 1, NotifLoginSuccess, map[string]interface{}{})

	require.NoError(t, err)
	assert.Zero(t, repo.getUserCalls)
	assert.Zero(t, repo.createQueueCalls)
}

func TestQueueAddsDefaultsAndCreatesQueueItem(t *testing.T) {
	repo := &notificationRepoStub{
		user: &models.User{ID: 7, Username: "alice", Email: "alice@example.test"},
		prefs: &models.NotificationPreferences{
			LoginSuccess: true,
		},
	}
	email := &notificationEmailStub{configured: true, appURL: "https://ipam.example.test"}
	svc := &NotificationService{repo: repo, email: email}

	err := svc.Queue(context.Background(), 7, NotifLoginSuccess, map[string]interface{}{
		"IP":        "192.0.2.10",
		"Device":    "Firefox",
		"Timestamp": "2026-05-17T03:00:00Z",
	})

	require.NoError(t, err)
	require.NotNil(t, repo.queued)
	assert.Equal(t, int64(7), repo.queued.UserID)
	assert.Equal(t, "alice@example.test", repo.queued.Email)
	assert.Equal(t, NotifLoginSuccess, repo.queued.Template)
	assert.Equal(t, "alice", repo.queuedData["Username"])
	assert.Equal(t, "https://ipam.example.test", repo.queuedData["AppURL"])
	assert.Equal(t, 1, repo.getPrefsCalls)
	assert.Equal(t, 1, repo.countRecentCalls)
}

func TestQueueSkipsDisabledPreference(t *testing.T) {
	repo := &notificationRepoStub{
		user:  &models.User{ID: 7, Username: "alice", Email: "alice@example.test"},
		prefs: &models.NotificationPreferences{LoginSuccess: false},
	}
	svc := &NotificationService{repo: repo, email: &notificationEmailStub{configured: true}}

	err := svc.Queue(context.Background(), 7, NotifLoginSuccess, map[string]interface{}{})

	require.NoError(t, err)
	assert.Zero(t, repo.countRecentCalls)
	assert.Zero(t, repo.createQueueCalls)
}

func TestQueueSkipsNonCriticalRateLimitedNotification(t *testing.T) {
	repo := &notificationRepoStub{
		user:        &models.User{ID: 7, Username: "alice", Email: "alice@example.test"},
		prefs:       &models.NotificationPreferences{LoginSuccess: true},
		recentCount: 10,
	}
	svc := &NotificationService{repo: repo, email: &notificationEmailStub{configured: true}}

	err := svc.Queue(context.Background(), 7, NotifLoginSuccess, map[string]interface{}{})

	require.NoError(t, err)
	assert.Zero(t, repo.createQueueCalls)
}

func TestQueueCriticalNotificationBypassesPreferencesAndRateLimit(t *testing.T) {
	repo := &notificationRepoStub{
		user:        &models.User{ID: 7, Username: "alice", Email: "alice@example.test"},
		prefs:       &models.NotificationPreferences{PasswordChanged: false},
		recentCount: 10,
	}
	svc := &NotificationService{repo: repo, email: &notificationEmailStub{configured: true}}

	err := svc.Queue(context.Background(), 7, NotifPasswordChanged, map[string]interface{}{})

	require.NoError(t, err)
	assert.Zero(t, repo.getPrefsCalls)
	assert.Zero(t, repo.countRecentCalls)
	assert.Equal(t, 1, repo.createQueueCalls)
}

func TestProcessQueueSendsAndMarksSent(t *testing.T) {
	repo := &notificationRepoStub{
		pending: []*models.NotificationQueue{{
			ID:       11,
			Email:    "alice@example.test",
			Template: NotifPasswordChanged,
			Data:     `{"Username":"alice","Timestamp":"now","IP":"192.0.2.10"}`,
		}},
	}
	email := &notificationEmailStub{configured: true}
	svc := &NotificationService{repo: repo, email: email}

	svc.ProcessQueue(context.Background())

	require.Len(t, email.sent, 1)
	assert.Equal(t, "alice@example.test", email.sent[0].to)
	assert.Equal(t, "Your password has been changed", email.sent[0].subject)
	assert.Contains(t, email.sent[0].body, "Hello alice")
	assert.Equal(t, []int64{11}, repo.markSent)
	assert.Empty(t, repo.markFailed)
	assert.Equal(t, 1, repo.cleanupCalls)
}

func TestProcessQueueRetriesSendFailure(t *testing.T) {
	repo := &notificationRepoStub{
		pending: []*models.NotificationQueue{{
			ID:         12,
			Email:      "alice@example.test",
			Template:   NotifPasswordChanged,
			Data:       `{"Username":"alice","Timestamp":"now","IP":"192.0.2.10"}`,
			RetryCount: 1,
		}},
	}
	email := &notificationEmailStub{configured: true, sendErr: errors.New("smtp down")}
	svc := &NotificationService{repo: repo, email: email}

	svc.ProcessQueue(context.Background())

	require.Len(t, repo.markFailed, 1)
	assert.Equal(t, int64(12), repo.markFailed[0].id)
	assert.Equal(t, 2, repo.markFailed[0].retryCount)
	assert.Contains(t, repo.markFailed[0].errMsg, "smtp down")
	require.NotNil(t, repo.markFailed[0].nextRetryAt)
	assert.True(t, time.Until(*repo.markFailed[0].nextRetryAt) > 4*time.Minute)
}

func TestProcessQueueMarksTemplateFailurePermanent(t *testing.T) {
	repo := &notificationRepoStub{
		pending: []*models.NotificationQueue{{
			ID:       13,
			Email:    "alice@example.test",
			Template: "unknown",
			Data:     `{"Username":"alice"}`,
		}},
	}
	svc := &NotificationService{repo: repo, email: &notificationEmailStub{configured: true}}

	svc.ProcessQueue(context.Background())

	require.Len(t, repo.markFailed, 1)
	assert.Equal(t, int64(13), repo.markFailed[0].id)
	assert.Equal(t, 0, repo.markFailed[0].retryCount)
	assert.Nil(t, repo.markFailed[0].nextRetryAt)
	assert.Contains(t, repo.markFailed[0].errMsg, "unknown notification template")
}

func TestProcessQueueUsesEmptyDataForMalformedJSON(t *testing.T) {
	repo := &notificationRepoStub{
		pending: []*models.NotificationQueue{{
			ID:       14,
			Email:    "alice@example.test",
			Template: NotifRequestSubmitted,
			Data:     `{bad json`,
		}},
	}
	email := &notificationEmailStub{configured: true}
	svc := &NotificationService{repo: repo, email: email}

	svc.ProcessQueue(context.Background())

	require.Len(t, email.sent, 1)
	assert.Equal(t, "New <no value> request submitted", email.sent[0].subject)
	assert.True(t, strings.Contains(email.sent[0].body, "Request ID: <no value>"))
}
