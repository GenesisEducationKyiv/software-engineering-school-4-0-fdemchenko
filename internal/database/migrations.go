package database

import (
	"database/sql"
	"errors"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func AutoMigrate(db *sql.DB, migrationsFS fs.FS, migratinsPath string, dbName string, closeConn bool) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	source, err := iofs.New(migrationsFS, migratinsPath)
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", source, dbName, driver)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	if closeConn {
		serr, derr := m.Close()
		if serr != nil {
			return serr
		}
		return derr
	}
	return nil
}
