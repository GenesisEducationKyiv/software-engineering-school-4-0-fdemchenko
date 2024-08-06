package services

import (
	"slices"
	"testing"

	"github.com/fdemchenko/exchanger/internal/repositories"
	"github.com/stretchr/testify/assert"
)

type SubscriptonsRepositoryMock struct {
	emails []string
}

func (er *SubscriptonsRepositoryMock) GetAll() ([]string, error) {
	return er.emails, nil
}

func (er *SubscriptonsRepositoryMock) Insert(email string) (int, error) {
	if slices.Contains(er.emails, email) {
		return 0, repositories.ErrDuplicateEmail
	}
	er.emails = append(er.emails, email)
	return 0, nil
}

func (er *SubscriptonsRepositoryMock) DeleteByEmail(email string) error {
	return nil
}

func (er *SubscriptonsRepositoryMock) DeleteByID(id int) error {
	return nil
}
func TestEmailService_CreateEmails(t *testing.T) {
	emailRepo := new(SubscriptonsRepositoryMock)
	emails := []string{"example@mail.com", "school@edu.ua"}

	emailService := NewSubscriptionService(emailRepo)
	for _, newEmail := range emails {
		_, err := emailService.Create(newEmail)
		assert.NoError(t, err)
	}
}

func TestEmailService_CaseInsensitiveness(t *testing.T) {
	emailRepo := new(SubscriptonsRepositoryMock)
	emails := []string{"example@mail.com", "EXamPlE@maIl.Com"}

	emailService := NewSubscriptionService(emailRepo)
	_, err := emailService.Create(emails[0])
	assert.NoError(t, err)

	_, err = emailService.Create(emails[1])
	assert.ErrorIs(t, err, repositories.ErrDuplicateEmail)
}

func TestEmailService_GetEmails(t *testing.T) {
	emailRepo := new(SubscriptonsRepositoryMock)
	emails := []string{"example@mail.com", "school@edu.ua"}

	emailService := NewSubscriptionService(emailRepo)
	for _, newEmail := range emails {
		_, err := emailService.Create(newEmail)
		assert.NoError(t, err)
	}

	emailsReturned, err := emailService.GetAll()
	assert.NoError(t, err)
	assert.ElementsMatch(t, emails, emailsReturned)
}

func TestEmailService_CreateDuplicateEmail(t *testing.T) {
	emailRepo := new(SubscriptonsRepositoryMock)
	emails := []string{"example@mail.com", "school@edu.ua"}

	emailService := NewSubscriptionService(emailRepo)
	for _, newEmail := range emails {
		_, err := emailService.Create(newEmail)
		assert.NoError(t, err)
	}

	_, err := emailService.Create(emails[0])
	assert.Equal(t, err, repositories.ErrDuplicateEmail)
}
