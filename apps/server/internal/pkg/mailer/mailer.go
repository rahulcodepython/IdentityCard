package mailer

import (
    "context"
    "fmt"
    "io"

    "github.com/resend/resend-go/v2"
    "gopkg.in/gomail.v2"

    "identitycard-server/internal/config"
)

type Mailer struct {
    resendClient *resend.Client
    from         string
    dialer       *gomail.Dialer
}

func New(cfg *config.Config) *Mailer {
    var client *resend.Client
    if cfg.ResendAPIKey != "" {
        client = resend.NewClient(cfg.ResendAPIKey)
    }

    from := cfg.ResendFrom
    if from == "" {
        from = cfg.SMTPFrom
    }

    dialer := gomail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword)

    return &Mailer{
        resendClient: client,
        from:         from,
        dialer:       dialer,
    }
}

type Attachment struct {
    Filename    string
    ContentType string
    Data        []byte
}

type BatchMessage struct {
    To          string
    Subject     string
    Body        string
    Attachments []Attachment
}

// Send sends a single email with attachments via Resend if configured, or falls back to SMTP.
func (m *Mailer) Send(ctx context.Context, to, subject, body string, attachments ...Attachment) error {
    if m.resendClient != nil {
        var resendAtts []*resend.Attachment
        for _, a := range attachments {
            resendAtts = append(resendAtts, &resend.Attachment{
                Content:     a.Data,
                Filename:    a.Filename,
                ContentType: a.ContentType,
            })
        }

        req := &resend.SendEmailRequest{
            From:        m.from,
            To:          []string{to},
            Subject:     subject,
            Text:        body,
            Attachments: resendAtts,
        }

        _, err := m.resendClient.Emails.SendWithContext(ctx, req)
        return err
    }

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

// SendBatch dispatches a bunch of emails to Resend using its Batch API (up to 100 per call).
func (m *Mailer) SendBatch(ctx context.Context, messages []BatchMessage) error {
    if len(messages) == 0 {
        return nil
    }

    if m.resendClient != nil {
        const batchLimit = 100
        for i := 0; i < len(messages); i += batchLimit {
            end := i + batchLimit
            if end > len(messages) {
                end = len(messages)
            }
            chunk := messages[i:end]

            var batchReqs []*resend.SendEmailRequest
            for _, msg := range chunk {
                var resendAtts []*resend.Attachment
                for _, a := range msg.Attachments {
                    resendAtts = append(resendAtts, &resend.Attachment{
                        Content:     a.Data,
                        Filename:    a.Filename,
                        ContentType: a.ContentType,
                    })
                }

                batchReqs = append(batchReqs, &resend.SendEmailRequest{
                    From:        m.from,
                    To:          []string{msg.To},
                    Subject:     msg.Subject,
                    Text:        msg.Body,
                    Attachments: resendAtts,
                })
            }

            if _, err := m.resendClient.Batch.SendWithContext(ctx, batchReqs); err != nil {
                return fmt.Errorf("resend batch send failed: %w", err)
            }
        }
        return nil
    }

    // Fallback: send one by one over SMTP
    for _, msg := range messages {
        if err := m.Send(ctx, msg.To, msg.Subject, msg.Body, msg.Attachments...); err != nil {
            return err
        }
    }
    return nil
}
