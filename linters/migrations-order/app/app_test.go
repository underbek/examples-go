package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	type args struct {
		migPath     string
		repoPath    string
		repoMigPath string
		revision    string
	}

	validArgs := args{
		migPath:     "testdata/migrations",
		repoPath:    "../../..",
		repoMigPath: "linters/migrations-order/app/testdata/migrations",
		revision:    "HEAD",
	}

	tests := []struct {
		name             string
		args             args
		newMigrationName string
		wantErr          string
	}{
		{
			name: "Unknown migrations path",
			args: args{
				migPath: "testdata/unknown",
			},
			wantErr: "failed to get current migrations: open testdata/unknown: no such file or directory",
		},
		{
			name: "Unknown repository",
			args: args{
				migPath:  "testdata/migrations",
				repoPath: "unknown",
			},
			wantErr: "failed to get revision migrations: repository does not exist",
		},
		{
			name: "Unknown repository revision",
			args: args{
				migPath:  "testdata/migrations",
				repoPath: "../../..",
				revision: "unknown",
			},
			wantErr: "failed to get revision migrations: reference not found",
		},
		{
			name: "Unknown repository migrations path",
			args: args{
				migPath:     "testdata/migrations",
				repoPath:    "../../..",
				repoMigPath: "linters/migrations-order/app/testdata/unknown",
				revision:    "HEAD",
			},
			wantErr: "failed to get revision migrations: directory not found",
		},
		{
			name:             "Invalid migrations order",
			args:             validArgs,
			newMigrationName: "20230903170057.sql",
			wantErr:          "current order violates migrations history",
		},
		{
			name:             "Future migration version",
			args:             validArgs,
			newMigrationName: time.Now().Add(time.Minute).Format(migrationTemplate),
			wantErr:          "future migrations is not permitted",
		},
		{
			name:             "Happy path",
			args:             validArgs,
			newMigrationName: "20230903170059.sql",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.newMigrationName != "" {
				fileName := filepath.Join("testdata/migrations/", tt.newMigrationName)

				f, err := os.Create(fileName) // nolint:gosec
				require.NoError(t, err)
				require.NoError(t, f.Close())

				defer func() {
					require.NoError(t, os.Remove(fileName))
				}()
			}

			err := Run(tt.args.migPath, tt.args.repoPath, tt.args.repoMigPath, tt.args.revision)
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}
