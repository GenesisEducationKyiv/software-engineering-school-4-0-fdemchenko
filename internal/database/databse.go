package database

import "database/sql"

type Options struct {
	MaxOpenConnections int
}

func OpenDB(dsn string, options Options) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(options.MaxOpenConnections)
	return db, nil
}
