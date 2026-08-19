// Package mailer sends transactional email (currently just the ID card
// PDF on publish/resend) over plain SMTP. Point it at Mailhog locally
// (see infra/docker-compose.yml) — mail lands at http://localhost:8025
// instead of a real inbox.
package mailer

import (
	"io"

	"gopkg.in/gomail.v2"

	"identitycard-server/internal/config"
)

type Mailer struct {
	dialer *gomail.Dialer
	from   string
}

func New(cfg *config.Config) *Mailer {
	dialer := gomail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword)
	return &Mailer{dialer: dialer, from: cfg.SMTPFrom}
}

type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

func (m *Mailer) Send(to, subject, body string, attachments ...Attachment) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/plain", body)

	for _, a := range attachments {
		data := a.Data
		msg.Attach(a.Filename,
			gomail.SetCopyFunc(func(w io.Writer) error {
				_, err := w.Write(data)
				return err
			}),
			gomail.SetHeader(map[string][]string{"Content-Type": {a.ContentType}}),
		)
	}

	return m.dialer.DialAndSend(msg)
}
