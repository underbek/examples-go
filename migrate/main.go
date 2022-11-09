package main

import (
	"fmt"
	"log"

	// need one of (pgx the better)
	_ "github.com/jackc/pgx/v4/stdlib" //pgx
	_ "github.com/lib/pq"              //libpq
)

const (
	host     = "localhost"
	port     = 5432
	user     = "user"
	password = "password"
	dbName   = "example_database"
)

func main() {
	connURL := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbName)

	if err := Run(connURL, WithFs(migrationsPath)); err != nil {
		log.Fatalf("failed executing migrate DB: %v\n", err)
	}
}
