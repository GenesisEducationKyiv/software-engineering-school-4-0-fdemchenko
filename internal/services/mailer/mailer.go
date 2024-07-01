package mailer

import (
	"bytes"
	"context"
	"math"
	"sync"
	"text/template"
	"time"

	"github.com/fdemchenko/exchanger/web/templates"
	"github.com/go-mail/mail/v2"
	"github.com/rs/zerolog/log"
)

const (
	DialerTimeout = 5 * time.Second
)

type mailer struct {
	dialer             *mail.Dialer
	sender             string
	emailService       EmailService
	rateService        RateService
	stopChan           chan bool
	connectionPoolSize int
	currencyTemplates  map[string]string
}

type RateService interface {
	GetRate(context.Context, string) (float32, error)
}

type EmailService interface {
	Create(email string) error
	GetAll() ([]string, error)
}

type MailerConfig struct {
	Host                       string
	Port                       int
	ConnectionPoolSize         int
	Username, Password, Sender string
	UpdateInterval             time.Duration
}

func NewMailerService(
	cfg MailerConfig,
	emailService EmailService,
	rateService RateService,
) *mailer {
	dialer := mail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)
	dialer.Timeout = DialerTimeout

	cfg.ConnectionPoolSize = int(math.Min(MaxConcurrentSMTPConn, float64(cfg.ConnectionPoolSize)))

	return &mailer{
		dialer:             dialer,
		sender:             cfg.Sender,
		emailService:       emailService,
		rateService:        rateService,
		stopChan:           make(chan bool),
		connectionPoolSize: cfg.ConnectionPoolSize,
		currencyTemplates:  make(map[string]string),
	}
}

func (m *mailer) StartEmailSending(updateInterval time.Duration) {
	ticker := time.NewTicker(updateInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				err := m.sendEmails()
				if err != nil {
					log.Error().Err(err).Send()
				}
			case <-m.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (m *mailer) StopEmailSending() {
	m.stopChan <- true
}

func (m *mailer) sendEmails() error {
	rate, err := m.rateService.GetRate(context.Background(), "usd")
	if err != nil {
		return err
	}
	err = m.executeCurrencyTemplates(rate)
	if err != nil {
		return err
	}
	emails, err := m.emailService.GetAll()
	if err != nil {
		return err
	}

	messagesChan := make(chan *mail.Message)
	errorsChan := make(chan error)
	wg := &sync.WaitGroup{}

	// setup connection pool.
	for i := 0; i < m.connectionPoolSize; i++ {
		wg.Add(1)
		go emailWorker(messagesChan, errorsChan, m.dialer, wg)
	}
	go func() {
		wg.Wait()
		close(errorsChan)
	}()

	go func() {
		for err := range errorsChan {
			log.Error().Err(err).Send()
		}
	}()

	for _, email := range emails {
		messagesChan <- m.createCurrencyUpdateMessage(email)
	}
	close(messagesChan)
	return nil
}

func (m *mailer) createCurrencyUpdateMessage(to string) *mail.Message {
	message := mail.NewMessage()
	message.SetHeader("From", m.sender)
	message.SetHeader("To", to)
	message.SetHeader("Subject", m.currencyTemplates["subject"])
	message.SetBody("text/plain", m.currencyTemplates["plainBody"])
	message.AddAlternative("text/html", m.currencyTemplates["htmlBody"])
	return message
}

func (m *mailer) executeCurrencyTemplates(data float32) error {
	tmpl, err := template.New("email").Parse(templates.MessageTemplate)
	if err != nil {
		return err
	}
	templateNames := []string{"subject", "plainBody", "htmlBody"}
	templateBuffer := new(bytes.Buffer)
	for _, templatesName := range templateNames {
		err = tmpl.ExecuteTemplate(templateBuffer, templatesName, data)
		if err != nil {
			return err
		}
		m.currencyTemplates[templatesName] = templateBuffer.String()
		templateBuffer.Reset()
	}
	return nil
}
