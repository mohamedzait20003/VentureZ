package mailer

import (
	"context"
	"fmt"
	"net"
	"net/smtp"
	"strings"
)

type SMTPMailer struct {
	addr string
	host string
	auth smtp.Auth
	from string
}

func NewSMTP(host, port, username, pass, from string) *SMTPMailer {
	return &SMTPMailer{
		addr: net.JoinHostPort(host, port),
		host: host,
		auth: smtp.PlainAuth("", username, pass, host),
		from: from,
	}
}

func (m *SMTPMailer) Send(_ context.Context, to, subject, htmlBody, textBody string) error {
	return smtp.SendMail(m.addr, m.auth, m.from, []string{to}, buildMIME(m.from, to, subject, textBody, htmlBody))
}

func buildMIME(from, to, subject, text, html string) []byte {
	const boundary = "venturez-alt-boundary-9f2c1a"
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", to)
	fmt.Fprintf(&b, "Subject: %s\r\n", subject)
	b.WriteString("MIME-Version: 1.0\r\n")
	fmt.Fprintf(&b, "Content-Type: multipart/alternative; boundary=%q\r\n\r\n", boundary)

	fmt.Fprintf(&b, "--%s\r\n", boundary)
	b.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n")
	b.WriteString(text)
	b.WriteString("\r\n\r\n")

	fmt.Fprintf(&b, "--%s\r\n", boundary)
	b.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n\r\n")
	b.WriteString(html)
	b.WriteString("\r\n\r\n")

	fmt.Fprintf(&b, "--%s--\r\n", boundary)
	return []byte(b.String())
}
