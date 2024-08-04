package repositories

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/lib/pq"
)

type PostgresSubscriptionRepository struct {
	DB *sql.DB
}

func (em *PostgresSubscriptionRepository) Insert(email string) (int, error) {
	stmt := `INSERT INTO subscriptions (email) VALUES ($1) RETURNING id`

	var id int
	row := em.DB.QueryRow(stmt, email)
	if row.Err() != nil {
		var pgError *pq.Error
		if errors.As(row.Err(), &pgError) {
			// unique constraint error
			if pgError.Code == pq.ErrorCode(PostgreSQLUniqueViolationErrorCode) &&
				strings.Contains(pgError.Message, "subscriptions_email_key") {
				return 0, ErrDuplicateEmail
			}
		}
		return 0, row.Err()
	}
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (em *PostgresSubscriptionRepository) GetAll() ([]string, error) {
	query := `SELECT email FROM subscriptions`
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

func (em *PostgresSubscriptionRepository) DeleteByEmail(email string) error {
	query := `DELETE FROM subscriptions WHERE email = $1`

	result, err := em.DB.Exec(query, email)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrEmailDoesNotExist
	}
	return nil
}

func (em *PostgresSubscriptionRepository) DeleteByID(id int) error {
	query := `DELETE FROM subscriptions WHERE id = $1`

	result, err := em.DB.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrEmailDoesNotExist
	}
	return nil
}
