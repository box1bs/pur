package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectToDB(dsn string) (*Postgres, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &Postgres{DB: db}, nil
}