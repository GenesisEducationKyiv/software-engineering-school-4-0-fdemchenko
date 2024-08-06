package services

import "strings"

type EmailRepository interface {
	Insert(email string) error
	GetAll() ([]string, error)
}

type emailServiceImpl struct {
	emailRepository EmailRepository
}

func NewEmailService(emailRepository EmailRepository) *emailServiceImpl {
	return &emailServiceImpl{emailRepository: emailRepository}
}

func (es *emailServiceImpl) Create(email string) error {
	// email is case insensitive
	email = strings.ToLower(email)
	return es.emailRepository.Insert(email)
}

func (es *emailServiceImpl) GetAll() ([]string, error) {
	return es.emailRepository.GetAll()
}
