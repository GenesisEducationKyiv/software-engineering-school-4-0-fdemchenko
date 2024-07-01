package mailer

import (
	"sync"

	"github.com/go-mail/mail/v2"
)

const (
	MaxConcurrentSMTPConn = 10
)

func emailWorker(jobs chan *mail.Message, errors chan error, dialer *mail.Dialer, wg *sync.WaitGroup) {
	defer wg.Done()
	sender, err := dialer.Dial()
	if err != nil {
		errors <- err
		return
	}

	for message := range jobs {
		err := mail.Send(sender, message)
		if err != nil {
			errors <- err
		}
	}

	err = sender.Close()
	if err != nil {
		errors <- err
	}
}
