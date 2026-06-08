package mailer

import (
	"bytes"
	"html/template"
)

var (
	verifyTmpl = template.Must(template.New("verify").Parse(
		`<p>Welcome to VentureZ.</p>
		<p>Please confirm your email by clicking the link below:</p>
		<p><a href="{{.Link}}">Verify my email</a></p>
		<p>This link expires in 24 hours.</p>`))

	resetTmpl = template.Must(template.New("reset").Parse(
		`<p>We received a request to reset your VentureZ password.</p>
		<p><a href="{{.Link}}">Reset my password</a></p>
		<p>This link expires in 1 hour. If you didn't request this, you can ignore this email.</p>`))
)

func VerifyEmail(link string) (subject, html, text string) {
	var buf bytes.Buffer
	_ = verifyTmpl.Execute(&buf, struct{ Link string }{link})
	text = "Welcome to VentureZ. Verify your email: " + link + "\n(This link expires in 24 hours.)"
	return "Verify your VentureZ email", buf.String(), text
}

func PasswordReset(link string) (subject, html, text string) {
	var buf bytes.Buffer
	_ = resetTmpl.Execute(&buf, struct{ Link string }{link})
	text = "Reset your VentureZ password: " + link + "\n(This link expires in 1 hour. Ignore if you didn't request it.)"
	return "Reset your VentureZ password", buf.String(), text
}
