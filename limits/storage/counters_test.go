package storage

import (
	"context"
	"time"
)

func (s *TestSuite) Test_storage_CleanupCounters() {
	getCount := func() int {
		var count int
		s.NoError(s.db.QueryRow(context.Background(), "SELECT COUNT(1) FROM counters").Scan(&count))
		return count
	}

	tests := []struct {
		name             string
		outdateInterval  time.Duration
		wantRemovedCount int
	}{
		{
			name:             "Happy path for 24h",
			outdateInterval:  time.Hour * 24,
			wantRemovedCount: 12,
		},
		{
			name:             "Happy path for 1h",
			outdateInterval:  time.Minute * 90,
			wantRemovedCount: 14,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			initialCount := getCount()
			s.Greater(initialCount, 0)

			s.NoError(s.storage.CleanupCounters(context.Background(), tt.outdateInterval))

			resultCount := getCount()
			s.Equal(tt.wantRemovedCount, initialCount-resultCount)

			s.NoError(s.storage.CleanupCounters(context.Background(), tt.outdateInterval))
			s.Equal(resultCount, getCount())
		})
	}
}
