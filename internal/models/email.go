package models

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/lib/pq"
)

type EmailModel struct {
	DB *sql.DB
}

func (em *EmailModel) Insert(email string) error {
	stmt := `INSERT INTO emails (email) VALUES ($1)`

	_, err := em.DB.Exec(stmt, email)
	if err != nil {
		var pgError *pq.Error
		if errors.As(err, &pgError) {
			// unique constraint error
			if pgError.Code == "23505" && strings.Contains(pgError.Message, "emails_email_key") {
				return ErrDuplicateEmail
			}
		}
		return err
	}
	return nil
}

func (em *EmailModel) GetAll() ([]string, error) {
	query := `SELECT email FROM emails`
	var emails []string

	rows, err := em.DB.Query(query)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var currentEmail string
		err := rows.Scan(&currentEmail)
		if err != nil {
			return nil, err
		}
		emails = append(emails, currentEmail)
	}

	return emails, nil
}
