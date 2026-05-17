package services

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
)

type EmailService struct {
	configSvc *ConfigService
}

func NewEmailService(configSvc *ConfigService) *EmailService {
	return &EmailService{configSvc: configSvc}
}

type smtpConfig struct {
	host     string
	port     int
	username string
	password string
	from     string
	useTLS   bool
}

func (e *EmailService) loadSMTPConfig() (*smtpConfig, error) {
	host, _ := e.configSvc.Get("smtp_host")
	if host == "" {
		return nil, fmt.Errorf("SMTP not configured")
	}

	portStr, _ := e.configSvc.Get("smtp_port")
	port, err := strconv.Atoi(portStr)
	if err != nil || port == 0 {
		port = 587
	}

	username, _ := e.configSvc.Get("smtp_username")
	password, _ := e.configSvc.Get("smtp_password")
	from, _ := e.configSvc.Get("smtp_from")
	if from == "" {
		from = username
	}

	tlsStr, _ := e.configSvc.Get("smtp_tls")
	useTLS := tlsStr != "false"

	return &smtpConfig{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
		useTLS:   useTLS,
	}, nil
}

func (e *EmailService) Send(to, subject, body string) error {
	cfg, err := e.loadSMTPConfig()
	if err != nil {
		return err
	}

	msg := buildMessage(cfg.from, to, subject, body)
	addr := net.JoinHostPort(cfg.host, strconv.Itoa(cfg.port))

	var auth smtp.Auth
	if cfg.username != "" {
		auth = smtp.PlainAuth("", cfg.username, cfg.password, cfg.host)
	}

	if cfg.useTLS {
		tlsCfg := &tls.Config{ServerName: cfg.host}
		conn, err := tls.Dial("tcp", addr, tlsCfg)
		if err != nil {
			return fmt.Errorf("TLS dial: %w", err)
		}
		client, err := smtp.NewClient(conn, cfg.host)
		if err != nil {
			return fmt.Errorf("SMTP client: %w", err)
		}
		defer client.Close()

		if auth != nil {
			if err := client.Auth(auth); err != nil {
				return fmt.Errorf("SMTP auth: %w", err)
			}
		}
		if err := client.Mail(cfg.from); err != nil {
			return err
		}
		if err := client.Rcpt(to); err != nil {
			return err
		}
		w, err := client.Data()
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(msg))
		if err != nil {
			return err
		}
		return w.Close()
	}

	return smtp.SendMail(addr, auth, cfg.from, []string{to}, []byte(msg))
}

func (e *EmailService) TestConnection() error {
	cfg, err := e.loadSMTPConfig()
	if err != nil {
		return err
	}

	addr := net.JoinHostPort(cfg.host, strconv.Itoa(cfg.port))
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("cannot reach SMTP server: %w", err)
	}
	if err := conn.Close(); err != nil {
		return fmt.Errorf("closing SMTP test connection: %w", err)
	}
	return nil
}

func buildMessage(from, to, subject, body string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("From: %s\r\n", from))
	sb.WriteString(fmt.Sprintf("To: %s\r\n", to))
	sb.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	sb.WriteString("MIME-Version: 1.0\r\n")
	sb.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	sb.WriteString("\r\n")
	sb.WriteString(body)
	return sb.String()
}

func (e *EmailService) SendVerificationEmail(to, username, token string) error {
	appURL, _ := e.configSvc.Get("app_url")
	if appURL == "" {
		appURL = "http://localhost:3000"
	}
	subject := "Verify your email address"
	body := fmt.Sprintf(`Hello %s,

Thank you for registering. Please verify your email address by clicking the link below:

%s/verify-email?token=%s

This link will expire in 24 hours.

If you did not create an account, please ignore this email.
`, username, appURL, token)
	return e.Send(to, subject, body)
}

func (e *EmailService) SendApprovalRequestEmail(adminEmail, username, userEmail string) error {
	subject := "New user registration requires approval"
	body := fmt.Sprintf(`A new user has registered and requires your approval.

Username: %s
Email: %s

Please log in to the admin panel to approve or reject this registration.
`, username, userEmail)
	return e.Send(adminEmail, subject, body)
}

func (e *EmailService) SendApprovedEmail(to, username string) error {
	subject := "Your account has been approved"
	body := fmt.Sprintf(`Hello %s,

Your account has been approved. You can now log in.
`, username)
	return e.Send(to, subject, body)
}

func (e *EmailService) SendRejectedEmail(to, username, reason string) error {
	subject := "Your registration was not approved"
	body := fmt.Sprintf(`Hello %s,

Unfortunately, your registration was not approved.
`, username)
	if reason != "" {
		body += fmt.Sprintf("\nReason: %s\n", reason)
	}
	return e.Send(to, subject, body)
}

func (e *EmailService) SendAccountLockedEmail(to, username, unlockURL string, duration interface{}) error {
	subject := "Your account has been temporarily locked"
	body := fmt.Sprintf(`Hello %s,

Your account has been temporarily locked due to multiple failed login attempts.

To unlock your account immediately, click the link below:

%s

This link will expire in 24 hours. Your account will also unlock automatically after the lockout period ends.

If you did not attempt to log in, your password may be compromised. Please change it after unlocking your account.
`, username, unlockURL)
	return e.Send(to, subject, body)
}

func (e *EmailService) SendPasswordResetEmail(to, username, token string) error {
	appURL, _ := e.configSvc.Get("app_url")
	if appURL == "" {
		appURL = "http://localhost:3000"
	}
	subject := "Reset your password"
	body := fmt.Sprintf(`Hello %s,

A password reset was requested for your account. Click the link below to set a new password:

%s/reset-password?token=%s

This link expires in 1 hour. If you did not request a reset, you can ignore this email.
`, username, appURL, token)
	return e.Send(to, subject, body)
}

func (e *EmailService) SendFailedLoginAlertEmail(to, username, ipAddress string, count int) error {
	subject := "Multiple failed login attempts detected"
	body := fmt.Sprintf(`Hello %s,

We detected %d failed login attempts on your account from IP address %s.

If this was not you, your account may be under a brute force attack. We recommend:
- Changing your password
- Enabling two-factor authentication if not already enabled

If this was you, please verify your password and try again.
`, username, count, ipAddress)
	return e.Send(to, subject, body)
}
