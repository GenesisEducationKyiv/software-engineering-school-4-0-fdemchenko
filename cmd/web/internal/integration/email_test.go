package integration

import (
	"context"
	"database/sql"
	"testing"

	"github.com/fdemchenko/exchanger/cmd/web/internal/repositories"
	"github.com/fdemchenko/exchanger/cmd/web/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type EmailService interface {
	Create(email string) (int, error)
	GetAll() ([]string, error)
}

type EmailServiceSuite struct {
	suite.Suite
	emailService EmailService
	container    *postgres.PostgresContainer
}

func (em *EmailServiceSuite) SetupSuite() {
	t := em.T()
	container, err := CreateTestDBContainer()
	if err != nil {
		t.Fatal(err)
	}

	em.container = container
	dsn, err := container.ConnectionString(context.Background(), "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}

	repo := &repositories.PostgresSubscriptionRepository{DB: db}
	em.emailService = services.NewSubscriptionService(repo)
}

func (em *EmailServiceSuite) TearDownTest() {
	err := em.container.Restore(context.Background())
	if err != nil {
		em.T().Fatal(err)
	}
}

func (em *EmailServiceSuite) TestCreateEmail_Success() {
	_, err := em.emailService.Create("someemail@gmail.com")
	assert.NoError(em.T(), err)
}

func (em *EmailServiceSuite) TestCreateEmail_Duplicate() {
	t := em.T()
	_, err := em.emailService.Create("somemail@gmail.com")
	assert.NoError(t, err)

	_, err = em.emailService.Create("somemail@gmail.com")
	assert.ErrorIs(t, err, repositories.ErrDuplicateEmail)
}

func (em *EmailServiceSuite) TestGetEmails() {
	t := em.T()
	_, err := em.emailService.Create("somemail1@gmail.com")
	assert.NoError(t, err)

	_, err = em.emailService.Create("another@gmail.com")
	assert.NoError(t, err)

	emails, err := em.emailService.GetAll()
	assert.NoError(t, err)
	assert.ElementsMatch(t, emails, []string{"somemail1@gmail.com", "another@gmail.com"})
}

func (em *EmailServiceSuite) TearDownSuite() {
	if err := em.container.Terminate(context.Background()); err != nil {
		em.T().Fatal(err)
	}
}

func TestEmailSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping email service integrtion test...")
	}

	suite.Run(t, new(EmailServiceSuite))
}
