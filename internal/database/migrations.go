package database

import (
	"database/sql"
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs" // source of migrations
)

//go:embed migrations/*sql
var migrationsFS embed.FS

// closeDB is used primarily for tests, where leaking connection causing error of performing snapshot.
func AutoMigrate(db *sql.DB, closeDB bool) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	fs, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs", fs, "exchanger", driver)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	if closeDB {
		serr, derr := m.Close()
		if serr != nil {
			return serr
		}
		if derr != nil {
			return derr
		}
	}
	return nil
}
