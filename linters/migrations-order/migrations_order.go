package main

import (
	"errors"
	"flag"
	"log"
	"strings"

	"github.com/underbek/examples-go/linters/migrations-order/app"
	"github.com/underbek/examples-go/logger"
)

type args struct {
	RepoPath string `json:"repo"`
	MigPath  string `json:"path"`
	Revision string `json:"revision"`
}

func main() {
	l, err := logger.New(true)
	if err != nil {
		log.Fatalln("failed to init log: ", err)
	}
	l = l.Named("migrations-order")

	a := loadArgs()
	l = l.With("args", a)

	if err = validateArgs(a); err != nil {
		l.WithError(err).Fatal("invalid arguments")
	}

	if err := app.Run(a.MigPath, a.RepoPath, a.MigPath, a.Revision); err != nil {
		l.WithError(err).Fatal("failed")
	}

	l.Info("successfully passed")
}

func loadArgs() args {
	var a args

	flag.StringVar(&a.RepoPath, "repo", "", "path to repository from root dir")
	flag.StringVar(&a.MigPath, "path", "", "path to migrations from root dir")
	flag.StringVar(&a.Revision, "revision", "", "revision to compare")
	flag.Parse()

	return a
}

func validateArgs(a args) error {
	messages := make([]string, 0, 3)

	if a.RepoPath == "" {
		messages = append(messages, `"repo" must not be empty`)
	}

	if a.MigPath == "" {
		messages = append(messages, `"path" must not be empty`)
	}

	if a.Revision == "" {
		messages = append(messages, `"revision" must not be empty`)
	}

	if len(messages) > 0 {
		return errors.New(strings.Join(messages, "; "))
	}

	return nil
}
