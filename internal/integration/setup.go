package integration

import (
	"context"
	"database/sql"
	"time"

	"github.com/fdemchenko/exchanger/internal/database"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	LogOccurance          = 2
	ContainerStartTimeout = 10 * time.Second
)

func createTestDBContainer() (*postgres.PostgresContainer, error) {
	ctx := context.Background()

	dbName := "test"
	dbUser := "testuser"
	dbPassword := "postgres"

	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:16-alpine"),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(LogOccurance).
				WithStartupTimeout(ContainerStartTimeout)),
	)
	if err != nil {
		return nil, err
	}
	dsn, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	err = database.AutoMigrate(db, true)
	if err != nil {
		return nil, err
	}
	err = postgresContainer.Snapshot(context.Background(), postgres.WithSnapshotName("test-snapshot"))
	if err != nil {
		return nil, err
	}

	return postgresContainer, nil
}
