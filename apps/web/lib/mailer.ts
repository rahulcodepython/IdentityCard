import "server-only"

import nodemailer from "nodemailer"

// Node-side port of apps/server/internal/pkg/mailer — auth emails (OTP
// codes, TOTP-related notices) now originate from better-auth in this
// app, not the Go API, so they need their own SMTP client. Same Mailhog
// target locally (see infra/docker-compose.yml); sent mail shows up at
// http://localhost:8025.
const transport = nodemailer.createTransport({
    host: process.env.SMTP_HOST || "localhost",
    port: Number(process.env.SMTP_PORT || 1025),
    secure: false,
    auth: process.env.SMTP_USERNAME
        ? { user: process.env.SMTP_USERNAME, pass: process.env.SMTP_PASSWORD }
        : undefined,
})

const from = process.env.SMTP_FROM || "no-reply@identitycard.local"

export async function sendMail(to: string, subject: string, text: string) {
    await transport.sendMail({ from, to, subject, text })
}
