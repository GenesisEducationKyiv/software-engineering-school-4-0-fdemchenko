package main

import (
	"time"

	"github.com/go-mail/mail/v2"
)

const (
	MaxConcurrentSMTPConn = 10
	UnusedConnectionTime  = 30 * time.Second
)

func emailWorker(jobs chan *mail.Message, errors chan error, dialer *mail.Dialer) {
	var sender mail.SendCloser
	var err error
	open := false

	for {
		select {
		case m := <-jobs:
			if !open {
				if sender, err = dialer.Dial(); err != nil {
					errors <- err
				}
				open = true
			}
			if err := mail.Send(sender, m); err != nil {
				errors <- err
			}
		// Close the connection to the SMTP server if no email was sent in
		// the last 30 seconds.
		case <-time.After(UnusedConnectionTime):
			if open {
				if err := sender.Close(); err != nil {
					errors <- err
				}
				open = false
			}
		}
	}
}
