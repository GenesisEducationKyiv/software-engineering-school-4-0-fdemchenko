package data

import (
	"database/sql"
)

type CustomerPostgreSQLRepository struct {
	DB *sql.DB
}

func (ctr *CustomerPostgreSQLRepository) Insert(email string, subscriptionID int) (int, error) {
	query := `INSERT INTO customers (email, subscription_id) VALUES ($1, $2) RETURNING id`

	var id int
	row := ctr.DB.QueryRow(query, email, subscriptionID)
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
