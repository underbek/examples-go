package main

import (
	"database/sql"
	"fmt"
	"log"

	"errors"
	_ "github.com/lib/pq"

	"github.com/pressly/goose"
)

const (
	host          = "localhost"
	port          = 5432
	user          = "user"
	password      = "password"
	dbName        = "example_database"
	migrationPath = "migrate"
)

func main() {
	connURL := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbName)
	db, err := sql.Open("postgres", connURL)
	if err != nil {
		log.Fatalf("error connecting to db: %s", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("failed to close DB: %v\n", err)
		}
	}()

	err = goose.Up(db, migrationPath)
	if err != nil {
		log.Fatalf("failed executing migrate DB: %v\n", err)
	}
}

type Result struct {
	nominal int
	count int
}

func calc(amount int, nominals []int) (res []Result, err error) {
	for _, nominal := range nominals {
		count := amount / nominal
		if count == 0 {
			continue
		}
		res = append(res, Result{
			nominal: nominal,
			count: count,
		})

		amount = amount % nominal
	}

	if amount != 0 {
		return res, errors.New("Not fount nominals")
	}

	return res, nil
}
