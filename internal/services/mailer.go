package services

import (
	"bytes"
	"text/template"
	"time"

	"github.com/fdemchenko/exchanger/internal/models"
	"github.com/fdemchenko/exchanger/web/templates"
	"github.com/go-mail/mail/v2"
	"github.com/rs/zerolog/log"
)

const (
	DialerTimeout = 5 * time.Second
)

type Mailer struct {
	dialer         *mail.Dialer
	sender         string
	emailModel     *models.EmailModel
	rateService    *RateService
	updateInterval time.Duration
}

type MailerConfig struct {
	Host                       string
	Port                       int
	Username, Password, Sender string
	UpdateInterval             time.Duration
}

func NewMailerService(cfg MailerConfig, emailModel *models.EmailModel, rateService *RateService) Mailer {
	dialer := mail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	dialer.Timeout = DialerTimeout
	return Mailer{
		dialer:         dialer,
		sender:         cfg.Sender,
		emailModel:     emailModel,
		rateService:    rateService,
		updateInterval: cfg.UpdateInterval,
	}
}

func (m Mailer) StartBackgroundTask() {
	go func() {
		for range time.Tick(m.updateInterval) {
			rate, err := m.rateService.GetRate()
			if err != nil {
				log.Error().Err(err).Send()
				continue
			}
			msg, err := m.prepareMessage(rate)
			if err != nil {
				log.Error().Err(err).Send()
				continue
			}
			emails, err := m.emailModel.GetAll()
			if err != nil {
				log.Error().Err(err).Send()
				continue
			}

			for _, email := range emails {
				msg.SetHeader("To", email)
				err = m.dialer.DialAndSend(msg)
				if err != nil {
					log.Error().Err(err).Send()
				}
			}
		}
	}()
}

func (m Mailer) prepareMessage(data interface{}) (*mail.Message, error) {
	tmpl, err := template.New("email").Parse(templates.MessageTemplate)
	if err != nil {
		return nil, err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return nil, err
	}

	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return nil, err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return nil, err
	}

	msg := mail.NewMessage()
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	return msg, nil
}
