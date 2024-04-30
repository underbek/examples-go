package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/underbek/examples-go/limits/config"
	"github.com/underbek/examples-go/logger"
)

func TestScheduler_Run(t *testing.T) {
	l, err := logger.New(true)
	require.NoError(t, err)

	tests := []struct {
		name            string
		storageProvider func() storage
		cfg             config.Scheduler
		ctxTimeout      time.Duration
	}{
		{
			name: "Failed DB queries",
			storageProvider: func() storage {
				s := NewStorageMock(t)
				s.On("CleanupLimits", mock.Anything).Return(errors.New("any error")).Once()
				s.On("CleanupCounters", mock.Anything, time.Hour).Return(errors.New("any error")).Once()
				s.On("CleanupContext", mock.Anything, time.Hour).Return(errors.New("any error")).Once()
				return s
			},
			cfg: config.Scheduler{
				Cleanup: config.Cleanup{
					RunInterval:     time.Millisecond * 100,
					OutdateInterval: time.Hour,
				},
			},
			ctxTimeout: time.Millisecond * 50,
		},
		{
			name: "Happy path",
			storageProvider: func() storage {
				s := NewStorageMock(t)
				s.On("CleanupLimits", mock.Anything).Return(nil).Once()
				s.On("CleanupCounters", mock.Anything, time.Hour).Return(nil).Once()
				s.On("CleanupContext", mock.Anything, time.Hour).Return(nil).Once()
				return s
			},
			cfg: config.Scheduler{
				Cleanup: config.Cleanup{
					RunInterval:     time.Millisecond * 100,
					OutdateInterval: time.Hour,
				},
			},
			ctxTimeout: time.Millisecond * 50,
		},
		{
			name: "Happy path for 2 iterations",
			storageProvider: func() storage {
				s := NewStorageMock(t)
				s.On("CleanupLimits", mock.Anything).Return(nil).Times(2)
				s.On("CleanupCounters", mock.Anything, time.Hour).Return(nil).Times(2)
				s.On("CleanupContext", mock.Anything, time.Hour).Return(nil).Times(2)
				return s
			},
			cfg: config.Scheduler{
				Cleanup: config.Cleanup{
					RunInterval:     time.Millisecond * 100,
					OutdateInterval: time.Hour,
				},
			},
			ctxTimeout: time.Millisecond * 150,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.ctxTimeout)
			defer cancel()

			require.Equal(t, context.DeadlineExceeded, New(l, tt.storageProvider(), tt.cfg).Run(ctx))
		})
	}
}
