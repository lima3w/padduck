package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"text/template"
	"time"

	"ipam-next/models"
	"ipam-next/repository"
)

// Template name constants
const (
	NotifLoginSuccess    = "login-success"
	NotifLoginFailed     = "login-failed"
	NotifAccountLocked   = "account-locked"
	NotifPasswordChanged = "password-changed"
	NotifMFAEnabled      = "mfa-enabled"
	NotifMFADisabled     = "mfa-disabled"
	NotifAPITokenCreated = "api-token-created"
	NotifAPITokenRevoked = "api-token-revoked"
	NotifRoleChanged     = "role-changed"
	NotifSessionRevoked  = "session-revoked"
)

// criticalNotifications lists templates that bypass preferences and rate limits.
var criticalNotifications = map[string]bool{
	NotifAccountLocked:   true,
	NotifRoleChanged:     true,
	NotifPasswordChanged: true,
}

// notifTemplate holds the subject and body Go templates for a notification.
type notifTemplate struct {
	Subject string
	Body    string
}

// notifTemplates contains all 10 notification templates.
var notifTemplates = map[string]notifTemplate{
	NotifLoginSuccess: {
		Subject: "New login to your account",
		Body: `Hello {{.Username}},

A new login to your account was detected.

IP Address: {{.IP}}
Device:     {{.Device}}
Time:       {{.Timestamp}}

If this was not you, secure your account immediately by changing your password and reviewing your active sessions.

IPAM System`,
	},

	NotifLoginFailed: {
		Subject: "Failed login attempts detected",
		Body: `Hello {{.Username}},

We detected {{.Count}} failed login attempt(s) on your account.

IP Address: {{.IP}}
Time:       {{.Timestamp}}

If this was not you, your account may be under attack. We recommend changing your password and enabling two-factor authentication if you have not already done so.

If this was not you, secure your account immediately.

IPAM System`,
	},

	NotifAccountLocked: {
		Subject: "Your account has been locked",
		Body: `Hello {{.Username}},

Your account has been locked due to multiple failed login attempts.

IP Address: {{.IP}}
Time:       {{.Timestamp}}
Lockout Duration: {{.Duration}}

To unlock your account immediately, visit the link below:

{{.UnlockURL}}

This link will expire in 24 hours. Your account will also unlock automatically after the lockout period ends.

If this was not you, secure your account immediately by changing your password once unlocked.

IPAM System`,
	},

	NotifPasswordChanged: {
		Subject: "Your password has been changed",
		Body: `Hello {{.Username}},

Your account password was changed successfully.

Time: {{.Timestamp}}
IP Address: {{.IP}}

If this was not you, secure your account immediately by using the password reset link on the login page.

IPAM System`,
	},

	NotifMFAEnabled: {
		Subject: "Two-factor authentication enabled",
		Body: `Hello {{.Username}},

Two-factor authentication (2FA) has been enabled on your account.

Time: {{.Timestamp}}
IP Address: {{.IP}}

Your account now requires a one-time code in addition to your password when signing in.

If this was not you, secure your account immediately.

IPAM System`,
	},

	NotifMFADisabled: {
		Subject: "Two-factor authentication disabled",
		Body: `Hello {{.Username}},

Two-factor authentication (2FA) has been disabled on your account.

Time: {{.Timestamp}}
IP Address: {{.IP}}

Your account no longer requires a one-time code when signing in. If you did not make this change, re-enable 2FA immediately and change your password.

If this was not you, secure your account immediately.

IPAM System`,
	},

	NotifAPITokenCreated: {
		Subject: "New API token created",
		Body: `Hello {{.Username}},

A new API token has been created for your account.

Token Name: {{.TokenName}}
Time:       {{.Timestamp}}
IP Address: {{.IP}}

API tokens provide programmatic access to your account. If you did not create this token, revoke it immediately from your account settings.

If this was not you, secure your account immediately.

IPAM System`,
	},

	NotifAPITokenRevoked: {
		Subject: "API token revoked",
		Body: `Hello {{.Username}},

An API token associated with your account has been revoked.

Token Name: {{.TokenName}}
Reason:     {{.Reason}}
Time:       {{.Timestamp}}
IP Address: {{.IP}}

If you did not revoke this token, your account may have been accessed without your authorization.

If this was not you, secure your account immediately.

IPAM System`,
	},

	NotifRoleChanged: {
		Subject: "Your account role has been changed",
		Body: `Hello {{.Username}},

Your account role has been updated by an administrator.

Previous Role: {{.OldRole}}
New Role:      {{.NewRole}}
Changed By:    {{.ChangedBy}}
Time:          {{.Timestamp}}

This change affects what actions you can perform within the system. If you have questions about this change, please contact your administrator.

If this was not you, secure your account immediately.

IPAM System`,
	},

	NotifSessionRevoked: {
		Subject: "A session was signed out",
		Body: `Hello {{.Username}},

A session on your account was signed out.

Device:    {{.Device}}
Time:      {{.Timestamp}}

If you did not sign out this session, someone else may have access to your account. Change your password and review your active sessions immediately.

If this was not you, secure your account immediately.

IPAM System`,
	},
}

// NotificationService queues and delivers email notifications to users.
type NotificationService struct {
	repo  *repository.Repository
	email *EmailService
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(repo *repository.Repository, email *EmailService) *NotificationService {
	return &NotificationService{repo: repo, email: email}
}

// renderTemplate renders both the subject and body for the named template using data.
func (n *NotificationService) renderTemplate(tmplName string, data map[string]interface{}) (subject, body string, err error) {
	t, ok := notifTemplates[tmplName]
	if !ok {
		return "", "", fmt.Errorf("unknown notification template: %q", tmplName)
	}

	subjectTmpl, err := template.New("subject").Parse(t.Subject)
	if err != nil {
		return "", "", fmt.Errorf("parse subject template %q: %w", tmplName, err)
	}
	var subjectBuf bytes.Buffer
	if err = subjectTmpl.Execute(&subjectBuf, data); err != nil {
		return "", "", fmt.Errorf("render subject template %q: %w", tmplName, err)
	}

	bodyTmpl, err := template.New("body").Parse(t.Body)
	if err != nil {
		return "", "", fmt.Errorf("parse body template %q: %w", tmplName, err)
	}
	var bodyBuf bytes.Buffer
	if err = bodyTmpl.Execute(&bodyBuf, data); err != nil {
		return "", "", fmt.Errorf("render body template %q: %w", tmplName, err)
	}

	return subjectBuf.String(), bodyBuf.String(), nil
}

// preferenceEnabled returns true if the user's notification preferences allow
// the given template. It returns true for any unrecognized template name.
func preferenceEnabled(prefs *models.NotificationPreferences, tmplName string) bool {
	switch tmplName {
	case NotifLoginSuccess:
		return prefs.LoginSuccess
	case NotifLoginFailed:
		return prefs.LoginFailed
	case NotifAccountLocked:
		return prefs.AccountLocked
	case NotifPasswordChanged:
		return prefs.PasswordChanged
	case NotifMFAEnabled, NotifMFADisabled:
		return prefs.MFAChanges
	case NotifAPITokenCreated, NotifAPITokenRevoked:
		return prefs.APITokenChanges
	case NotifRoleChanged:
		return prefs.RoleChanges
	case NotifSessionRevoked:
		return prefs.SessionRevoked
	default:
		return true
	}
}

// Queue enqueues a notification for the given user. It respects notification
// preferences and a per-hour rate limit unless the template is critical.
func (n *NotificationService) Queue(ctx context.Context, userID int64, tmplName string, data map[string]interface{}) error {
	if !n.email.configSvc.IsSMTPConfigured() {
		return nil
	}
	// 1. Fetch user for email and username.
	user, err := n.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("notification queue: get user %d: %w", userID, err)
	}

	// Populate standard template variables if not already set.
	if _, ok := data["Username"]; !ok {
		data["Username"] = user.Username
	}
	appURL, _ := n.email.configSvc.Get("app_url")
	if appURL == "" {
		appURL = "http://localhost:3000"
	}
	if _, ok := data["AppURL"]; !ok {
		data["AppURL"] = appURL
	}

	critical := criticalNotifications[tmplName]

	if !critical {
		// 2. Check notification preferences.
		prefs, err := n.repo.GetNotificationPreferences(ctx, userID)
		if err != nil {
			return fmt.Errorf("notification queue: get preferences for user %d: %w", userID, err)
		}
		if !preferenceEnabled(prefs, tmplName) {
			return nil // silently skip
		}

		// 3. Rate-limit: no more than 10 notifications per hour per user.
		count, err := n.repo.CountRecentNotificationsSent(ctx, userID, time.Now().Add(-1*time.Hour))
		if err != nil {
			return fmt.Errorf("notification queue: count recent notifications for user %d: %w", userID, err)
		}
		if count >= 10 {
			log.Printf("[notification] rate limit reached for user %d (%s), skipping %s", userID, user.Username, tmplName)
			return nil
		}
	}

	// 4. Marshal data to JSON.
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("notification queue: marshal data: %w", err)
	}

	// 5. Enqueue.
	_, err = n.repo.CreateNotificationQueueItem(ctx, userID, user.Email, tmplName, string(dataJSON))
	return err
}

// retryDelay returns the next retry delay for the given retry count, or zero
// duration with a nil indicator when the notification should be permanently failed.
func retryDelay(retryCount int) *time.Duration {
	delays := []time.Duration{
		1 * time.Minute,
		5 * time.Minute,
		15 * time.Minute,
		1 * time.Hour,
		4 * time.Hour,
		24 * time.Hour,
	}
	if retryCount >= len(delays) {
		return nil
	}
	d := delays[retryCount]
	return &d
}

// ProcessQueue fetches pending notifications and attempts to deliver them.
func (n *NotificationService) ProcessQueue(ctx context.Context) {
	if !n.email.configSvc.IsSMTPConfigured() {
		return
	}
	items, err := n.repo.GetPendingNotifications(ctx, 50)
	if err != nil {
		log.Printf("[notification] get pending notifications: %v", err)
		return
	}

	for _, item := range items {
		// Unmarshal stored JSON data back into a map for rendering.
		var data map[string]interface{}
		if item.Data != "" {
			if jsonErr := json.Unmarshal([]byte(item.Data), &data); jsonErr != nil {
				log.Printf("[notification] unmarshal data for item %d: %v", item.ID, jsonErr)
				data = map[string]interface{}{}
			}
		} else {
			data = map[string]interface{}{}
		}

		subject, body, renderErr := n.renderTemplate(item.Template, data)
		if renderErr != nil {
			log.Printf("[notification] render template for item %d (%s): %v", item.ID, item.Template, renderErr)
			// Treat template errors as permanent failures.
			errMsg := renderErr.Error()
			if markErr := n.repo.MarkNotificationFailed(ctx, item.ID, errMsg, item.RetryCount, nil); markErr != nil {
				log.Printf("[notification] mark failed for item %d: %v", item.ID, markErr)
			}
			continue
		}

		sendErr := n.email.Send(item.Email, subject, body)
		if sendErr == nil {
			if markErr := n.repo.MarkNotificationSent(ctx, item.ID); markErr != nil {
				log.Printf("[notification] mark sent for item %d: %v", item.ID, markErr)
			}
			continue
		}

		log.Printf("[notification] send failed for item %d (%s → %s): %v", item.ID, item.Template, item.Email, sendErr)
		newRetryCount := item.RetryCount + 1
		delay := retryDelay(newRetryCount)
		var nextRetryAt *time.Time
		if delay != nil {
			t := time.Now().Add(*delay)
			nextRetryAt = &t
		}
		errMsg := sendErr.Error()
		if markErr := n.repo.MarkNotificationFailed(ctx, item.ID, errMsg, newRetryCount, nextRetryAt); markErr != nil {
			log.Printf("[notification] mark failed for item %d: %v", item.ID, markErr)
		}
	}

	// Clean up old processed notifications; ignore errors.
	if err := n.repo.CleanupOldNotifications(ctx); err != nil {
		log.Printf("[notification] cleanup old notifications: %v", err)
	}
}

// StartWorker launches a background goroutine that calls ProcessQueue every 30
// seconds until ctx is cancelled.
func (n *NotificationService) StartWorker(ctx context.Context) {
	if !n.email.configSvc.IsSMTPConfigured() {
		log.Printf("[notification] SMTP not configured — email delivery disabled")
	}
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				n.ProcessQueue(ctx)
			}
		}
	}()
}

// --- Service-level wrappers (called by handlers) ---

func (s *Service) GetNotificationPreferences(ctx context.Context, userID int64) (*models.NotificationPreferences, error) {
	return s.repository.GetNotificationPreferences(ctx, userID)
}

func (s *Service) UpsertNotificationPreferences(ctx context.Context, prefs *models.NotificationPreferences) (*models.NotificationPreferences, error) {
	return s.repository.UpsertNotificationPreferences(ctx, prefs)
}

func (s *Service) GetNotificationStats(ctx context.Context) (map[string]int64, error) {
	return s.repository.GetNotificationStats(ctx)
}
