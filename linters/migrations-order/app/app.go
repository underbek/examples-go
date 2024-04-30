package app

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

const migrationTemplate = "20060102150405.sql"

func Run(migPath, repoPath, repoMigPath, revision string) error {
	cms, err := getCurrent(migPath)
	if err != nil {
		return fmt.Errorf("failed to get current migrations: %w", err)
	}
	if len(cms) == 0 {
		return nil
	}

	//We add one day and cut off the hours in the current time so that the difference in time zones does not affect migrations
	now := time.Now().UTC().AddDate(0, 0, 1)
	currentDay := now.Truncate(24 * time.Hour)

	if strings.Compare(cms[len(cms)-1], currentDay.Format(migrationTemplate)) == 1 {
		return errors.New("future migrations is not permitted")
	}

	rms, err := getFromRevision(repoPath, repoMigPath, revision)
	if err != nil {
		return fmt.Errorf("failed to get revision migrations: %w", err)
	}
	if len(rms) == 0 {
		return nil
	}

	for i := range rms {
		if len(cms) > i && strings.Compare(rms[i], cms[i]) == 1 {
			return errors.New("current order violates migrations history")
		}
	}

	return nil
}

func getCurrent(migPath string) ([]string, error) {
	files, err := os.ReadDir(migPath)
	if err != nil {
		return nil, err
	}

	res := make([]string, len(files))
	for i := range files {
		res[i] = files[i].Name()
	}

	return res, nil
}

func getFromRevision(repoPath, migPath, revision string) ([]string, error) {
	c, err := getRevisionCommit(repoPath, revision)
	if err != nil {
		return nil, err
	}

	tree, err := c.Tree()
	if err != nil {
		return nil, err
	}

	tree, err = tree.Tree(migPath)
	if err != nil {
		return nil, err
	}

	res := make([]string, len(tree.Entries))
	for i := range tree.Entries {
		res[i] = tree.Entries[i].Name
	}

	return res, nil
}

func getRevisionCommit(repoPath, name string) (*object.Commit, error) {
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, err
	}

	h, err := r.ResolveRevision(plumbing.Revision(name))
	if err != nil {
		return nil, err
	}

	return r.CommitObject(*h)
}
