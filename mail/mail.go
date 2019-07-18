package mail

import (
	"net/smtp"
	"strconv"
	"strings"
)

type Config struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
	Server   string `toml:"server"`
	Port     int    `toml:"port"`
}

type Mailer struct {
	*Config
}

func NewMailer(c *Config) *Mailer {
	mailer := &Mailer{c}
	return mailer
}

func (m *Mailer) Send(to []string, nickname, subject, body string) (err error) {
	auth := smtp.PlainAuth("", m.Username, m.Password, m.Server)
	user := m.Username
	contentType := "Content-Type: text/plain; charset=UTF-8"
	msg := []byte("To: " + strings.Join(to, ",") + "\r\nFrom: " + nickname +
		"<" + user + ">\r\nSubject: " + subject + "\r\n" + contentType + "\r\n\r\n" + body)
	return smtp.SendMail(m.Server+":"+strconv.Itoa(m.Port), auth, user, to, msg)
}
