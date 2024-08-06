package services

import (
	"bytes"
	"math"
	"text/template"

	"github.com/fdemchenko/exchanger/cmd/mailer/internal/config"
	"github.com/fdemchenko/exchanger/web/templates"
	"github.com/go-mail/mail/v2"
	"github.com/rs/zerolog/log"
)

type MailerService struct {
	dialer            *mail.Dialer
	sender            string
	currencyTemplates map[string]string
	parsedTemplate    *template.Template
	jobsChan          chan *mail.Message
	errorsChan        chan error
}

func NewMailerService(
	cfg config.SMTPConfig,
) *MailerService {
	dialer := mail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)

	cfg.ConnectionPoolSize = int(math.Min(MaxConcurrentSMTPConn, float64(cfg.ConnectionPoolSize)))

	errorsChan := make(chan error)

	go func() {
		for err := range errorsChan {
			log.Error().Err(err).Send()
		}
	}()

	return &MailerService{
		dialer:            dialer,
		sender:            cfg.Sender,
		currencyTemplates: make(map[string]string),
		parsedTemplate:    template.Must(template.New("email").Parse(templates.MessageTemplate)),
		jobsChan:          make(chan *mail.Message),
		errorsChan:        errorsChan,
	}
}

func (ms *MailerService) UpdateCurrencyRateTemplates(rate float32) error {
	templateNames := []string{"subject", "plainBody", "htmlBody"}
	templateBuffer := new(bytes.Buffer)
	for _, templatesName := range templateNames {
		err := ms.parsedTemplate.ExecuteTemplate(templateBuffer, templatesName, rate)
		if err != nil {
			return err
		}
		ms.currencyTemplates[templatesName] = templateBuffer.String()
		templateBuffer.Reset()
	}
	return nil
}

func (ms *MailerService) StartWorkers(connectionPoolSize int) {
	for i := 0; i < connectionPoolSize; i++ {
		go emailWorker(ms.jobsChan, ms.errorsChan, ms.dialer)
	}
}

func (ms *MailerService) SendEmail(to string) {
	message := mail.NewMessage()
	message.SetHeader("From", ms.sender)
	message.SetHeader("To", to)
	message.SetHeader("Subject", ms.currencyTemplates["subject"])
	message.SetBody("text/plain", ms.currencyTemplates["plainBody"])
	message.AddAlternative("text/html", ms.currencyTemplates["htmlBody"])

	ms.jobsChan <- message
}
