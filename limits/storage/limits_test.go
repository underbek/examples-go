package storage

import (
	"context"
)

func (s *TestSuite) Test_storage_CleanupLimits() {
	getCount := func() int {
		var count int
		s.NoError(s.db.QueryRow(context.Background(), "SELECT COUNT(1) FROM limits").Scan(&count))
		return count
	}

	tests := []struct {
		name             string
		wantRemovedCount int
	}{
		{
			name:             "Happy path",
			wantRemovedCount: 3,
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			initialCount := getCount()
			s.Greater(initialCount, 0)

			s.NoError(s.storage.CleanupLimits(context.Background()))

			resultCount := getCount()
			s.Equal(tt.wantRemovedCount, initialCount-resultCount)

			s.NoError(s.storage.CleanupLimits(context.Background()))
			s.Equal(resultCount, getCount())
		})
	}
}
