package notify

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"
	"time"
)

// Config holds SMTP notification settings.
type Config struct {
	Enabled   bool
	SMTPHost  string
	SMTPPort  int
	From      string
	To        []string
	Username  string
	Password  string
}

// JobFailure represents a failed cron job event.
type JobFailure struct {
	JobName   string
	ExitCode  int
	Output    string
	OccurredAt time.Time
}

// Notifier sends failure notifications.
type Notifier struct {
	cfg    Config
	logger *slog.Logger
}

// New creates a new Notifier.
func New(cfg Config, logger *slog.Logger) *Notifier {
	return &Notifier{cfg: cfg, logger: logger}
}

// Notify sends an email notification for a job failure.
func (n *Notifier) Notify(failure JobFailure) error {
	if !n.cfg.Enabled {
		n.logger.Debug("notifications disabled, skipping", "job", failure.JobName)
		return nil
	}

	subject := fmt.Sprintf("[cronlog] Job failed: %s", failure.JobName)
	body := n.buildBody(failure)

	msg := n.buildMessage(subject, body)
	addr := fmt.Sprintf("%s:%d", n.cfg.SMTPHost, n.cfg.SMTPPort)

	var auth smtp.Auth
	if n.cfg.Username != "" {
		auth = smtp.PlainAuth("", n.cfg.Username, n.cfg.Password, n.cfg.SMTPHost)
	}

	if err := smtp.SendMail(addr, auth, n.cfg.From, n.cfg.To, msg); err != nil {
		n.logger.Error("failed to send notification", "job", failure.JobName, "err", err)
		return fmt.Errorf("notify: send mail: %w", err)
	}

	n.logger.Info("failure notification sent", "job", failure.JobName, "to", n.cfg.To)
	return nil
}

func (n *Notifier) buildBody(f JobFailure) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Job:        %s\n", f.JobName))
	sb.WriteString(fmt.Sprintf("Exit Code:  %d\n", f.ExitCode))
	sb.WriteString(fmt.Sprintf("Occurred:   %s\n", f.OccurredAt.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("\nOutput:\n%s\n", f.Output))
	return sb.String()
}

func (n *Notifier) buildMessage(subject, body string) []byte {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "From: %s\r\n", n.cfg.From)
	fmt.Fprintf(&buf, "To: %s\r\n", strings.Join(n.cfg.To, ", "))
	fmt.Fprintf(&buf, "Subject: %s\r\n", subject)
	fmt.Fprintf(&buf, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&buf, "Content-Type: text/plain; charset=UTF-8\r\n")
	fmt.Fprintf(&buf, "\r\n%s", body)
	return buf.Bytes()
}
