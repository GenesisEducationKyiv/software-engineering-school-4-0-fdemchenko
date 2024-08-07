package services

import (
	"strings"
)

type SubscriptonsRepository interface {
	Insert(email string) (int, error)
	GetAll() ([]string, error)
	DeleteByEmail(email string) error
	DeleteByID(id int) error
}

type subscriptionServiceImpl struct {
	subscriptionsRepository SubscriptonsRepository
}

func NewSubscriptionService(
	subscriptionRepository SubscriptonsRepository,
) *subscriptionServiceImpl {
	return &subscriptionServiceImpl{
		subscriptionsRepository: subscriptionRepository,
	}
}

func (ss *subscriptionServiceImpl) Create(email string) (int, error) {
	// email is case insensitive
	email = strings.ToLower(email)
	return ss.subscriptionsRepository.Insert(email)
}

func (ss *subscriptionServiceImpl) GetAll() ([]string, error) {
	return ss.subscriptionsRepository.GetAll()
}

func (ss *subscriptionServiceImpl) DeleteByEmail(email string) error {
	// email is case insensitive
	email = strings.ToLower(email)
	return ss.subscriptionsRepository.DeleteByEmail(email)
}

func (ss *subscriptionServiceImpl) DeleteByID(id int) error {
	return ss.subscriptionsRepository.DeleteByID(id)
}
