package service

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"github.com/shipping/notification-service/internal/config"
	"github.com/shipping/shared/pkg/logger"
)

type NotificationService struct{
	cfg *config.Config
}

func NewNotificationService(cfg *config.Config) *NotificationService {
	return &NotificationService{cfg: cfg}
}

// SendEmail sends a plain text email using configured SMTP server.
func (s *NotificationService) SendEmail(to, subject, body string) error {
	log := logger.Get()
	log.Info("sending email",
		logger.String("to", to),
		logger.String("subject", subject),
	)

	if s.cfg == nil {
		return fmt.Errorf("smtp config not loaded")
	}

	from := s.cfg.SMTPFrom
	if from == "" {
		from = s.cfg.SMTPUser
	}
	if from == "" {
		return fmt.Errorf("smtp from not configured")
	}

	header := make(map[string]string)
	header["From"] = from
	header["To"] = to
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""

	msg := ""
	for k, v := range header {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n" + body

	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)

	auth := smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPassword, s.cfg.SMTPHost)

	// Dial TCP
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("dial smtp: %w", err)
	}
	// ensure connection closed on exit
	defer func() {
		_ = conn.Close()
	}()

	c, err := smtp.NewClient(conn, s.cfg.SMTPHost)
	if err != nil {
		return fmt.Errorf("new smtp client: %w", err)
	}
	defer func() {
		if err := c.Quit(); err != nil {
			// best-effort log; do not override original error
			logger.Get().Error("smtp quit", logger.String("err", err.Error()))
		}
	}()

	// STARTTLS if available
	if ok, _ := c.Extension("STARTTLS"); ok {
		tlsCfg := &tls.Config{ServerName: s.cfg.SMTPHost}
		if err = c.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}

	// Auth
	if s.cfg.SMTPUser != "" && s.cfg.SMTPPassword != "" {
		if err = c.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err = c.Mail(from); err != nil {
		return fmt.Errorf("mail from: %w", err)
	}
	for _, addrTo := range strings.Split(to, ",") {
		if err = c.Rcpt(strings.TrimSpace(addrTo)); err != nil {
			return fmt.Errorf("rcpt to %s: %w", addrTo, err)
		}
	}

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("write msg: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("close data: %w", err)
	}

	log.Info("email sent", logger.String("to", to))
	return nil
}

func (s *NotificationService) SendSMS(phone, message string) error {
	logger.Info("sending SMS",
		logger.String("phone", phone),
	)
	// In production: integrate with Twilio
	fmt.Printf("[SMS] To: %s | Message: %s\n", phone, message)
	return nil
}
