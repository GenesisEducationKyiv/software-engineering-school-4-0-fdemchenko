package services

import (
	"slices"
	"testing"

	"github.com/fdemchenko/exchanger/internal/repositories"
	"github.com/stretchr/testify/assert"
)

type EmailRepositoryMock struct {
	emails []string
}

func (er *EmailRepositoryMock) GetAll() ([]string, error) {
	return er.emails, nil
}

func (er *EmailRepositoryMock) Insert(email string) error {
	if slices.Contains(er.emails, email) {
		return repositories.ErrDuplicateEmail
	}
	er.emails = append(er.emails, email)
	return nil
}

func (er *EmailRepositoryMock) Delete(email string) error {
	return nil
}

func TestEmailService_CreateEmails(t *testing.T) {
	emailRepo := new(EmailRepositoryMock)
	emails := []string{"example@mail.com", "school@edu.ua"}

	emailService := NewEmailService(emailRepo)
	for _, newEmail := range emails {
		err := emailService.Create(newEmail)
		assert.NoError(t, err)
	}
}

func TestEmailService_CaseInsensitiveness(t *testing.T) {
	emailRepo := new(EmailRepositoryMock)
	emails := []string{"example@mail.com", "EXamPlE@maIl.Com"}

	emailService := NewEmailService(emailRepo)
	err := emailService.Create(emails[0])
	assert.NoError(t, err)

	err = emailService.Create(emails[1])
	assert.ErrorIs(t, err, repositories.ErrDuplicateEmail)
}

func TestEmailService_GetEmails(t *testing.T) {
	emailRepo := new(EmailRepositoryMock)
	emails := []string{"example@mail.com", "school@edu.ua"}

	emailService := NewEmailService(emailRepo)
	for _, newEmail := range emails {
		err := emailService.Create(newEmail)
		assert.NoError(t, err)
	}

	emailsReturned, err := emailService.GetAll()
	assert.NoError(t, err)
	assert.ElementsMatch(t, emails, emailsReturned)
}

func TestEmailService_CreateDuplicateEmail(t *testing.T) {
	emailRepo := new(EmailRepositoryMock)
	emails := []string{"example@mail.com", "school@edu.ua"}

	emailService := NewEmailService(emailRepo)
	for _, newEmail := range emails {
		err := emailService.Create(newEmail)
		assert.NoError(t, err)
	}

	err := emailService.Create(emails[0])
	assert.Equal(t, err, repositories.ErrDuplicateEmail)
}
