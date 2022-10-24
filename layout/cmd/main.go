package main

import (
	"log"

	"layout/internal/app"
	"layout/internal/config"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	a, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(a.Run())
}
