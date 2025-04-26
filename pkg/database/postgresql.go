package database

import (
	"errors"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func PostgresGormConnection(conn string) (*gorm.DB, error) {

	if conn == "" {
		return nil, errors.New("empty connection string")
	}

	db, err := gorm.Open(postgres.Open(conn), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	return db, nil
}
