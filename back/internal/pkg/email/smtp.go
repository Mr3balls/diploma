package email

import (
	"fmt"
	"net/smtp"
)

// Sender sends emails via SMTP (Gmail port 587 / STARTTLS).
type Sender struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewSender(host, port, username, password, from string) *Sender {
	return &Sender{host: host, port: port, username: username, password: password, from: from}
}

func (s *Sender) Send(to, subject, htmlBody string) error {
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		s.from, to, subject, htmlBody,
	)

	return smtp.SendMail(s.host+":"+s.port, auth, s.from, []string{to}, []byte(msg))
}
