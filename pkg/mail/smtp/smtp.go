package smpt

import (
	"errors"
	"fmt"
	"net/smtp"
	"todoai/pkg/mail"
)

type Mail struct {
	From     string
	smtpPort int
	smtpHost string
	password string
}

func NewSMTPSender(from, pass, host string, port int) (*Mail, error) {
	if !mail.IsEmailValid(from) {
		return nil, errors.New("invalid from email")
	}

	return &Mail{From: from, smtpPort: port, smtpHost: host, password: pass}, nil
}

func (m *Mail) Send(input *mail.MailData) error {
	if err := input.Validate(); err != nil {
		return err
	}

	from := m.From
	password := m.password

	to := []string{input.To}

	msg := []byte("Subject: " + input.Subject + "\r\n" +
		"\r\n" +
		input.Body)

	auth := smtp.PlainAuth("", from, password, m.smtpHost)

	err := smtp.SendMail(m.smtpHost+":"+fmt.Sprintf("%d", m.smtpPort), auth, from, to, msg)
	if err != nil {
		return err
	}

	return nil
}
