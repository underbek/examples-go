package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/underbek/examples-go/limits/domain"
	"github.com/underbek/examples-go/utils"
)

func TestGenerateStartEndPeriod(t *testing.T) {
	lc, err := time.LoadLocation("Europe/Moscow")
	require.NoError(t, err)

	tests := []struct {
		name        string
		currentTime time.Time
		limit       domain.Limit
		startPeriod time.Time
		endPeriod   time.Time
		err         string
	}{
		{
			name:        "period is nil",
			currentTime: time.Date(2023, 5, 13, 15, 39, 0, 0, lc),
			limit: domain.Limit{
				ID:       11,
				Timezone: utils.ToPtr("UTC"),
			},
			err: "period is nil for limit 11",
		},
		{
			name:        "timezone is nil",
			currentTime: time.Date(2023, 5, 13, 15, 39, 0, 0, lc),
			limit: domain.Limit{
				ID:     11,
				Period: utils.ToPtr(domain.PeriodTypeCALENDARDAY),
			},
			err: "timezone is nil for limit 11",
		},
		{
			name:        "invalid timezone",
			currentTime: time.Date(2023, 5, 13, 15, 39, 0, 0, lc),
			limit: domain.Limit{
				ID:       11,
				Period:   utils.ToPtr(domain.PeriodTypeCALENDARDAY),
				Timezone: utils.ToPtr("INVALID"),
			},
			err: "load location failed for limit 11:",
		},
		{
			name:        "daily UTC",
			currentTime: time.Date(2023, 5, 13, 15, 39, 0, 0, lc),
			limit: domain.Limit{
				Period:   utils.ToPtr(domain.PeriodTypeCALENDARDAY),
				Timezone: utils.ToPtr("UTC"),
			},
			startPeriod: time.Date(2023, 5, 13, 0, 0, 0, 0, time.UTC),
			endPeriod:   time.Date(2023, 5, 14, 0, 0, 0, 0, time.UTC),
		},
		{
			name:        "daily London",
			currentTime: time.Date(2023, 5, 13, 15, 39, 0, 0, lc),
			limit: domain.Limit{
				Period:   utils.ToPtr(domain.PeriodTypeCALENDARDAY),
				Timezone: utils.ToPtr("Europe/London"),
			},
			startPeriod: time.Date(2023, 5, 12, 23, 0, 0, 0, time.UTC),
			endPeriod:   time.Date(2023, 5, 13, 23, 0, 0, 0, time.UTC),
		},
		{
			name:        "weekly London",
			currentTime: time.Date(2023, 5, 13, 15, 39, 0, 0, lc),
			limit: domain.Limit{
				Period:   utils.ToPtr(domain.PeriodTypeCALENDARWEEK),
				Timezone: utils.ToPtr("Europe/London"),
			},
			startPeriod: time.Date(2023, 5, 7, 23, 0, 0, 0, time.UTC),
			endPeriod:   time.Date(2023, 5, 14, 23, 0, 0, 0, time.UTC),
		},
		{
			name:        "monthly London",
			currentTime: time.Date(2023, 5, 13, 15, 39, 0, 0, lc),
			limit: domain.Limit{
				Period:   utils.ToPtr(domain.PeriodTypeCALENDARMONTH),
				Timezone: utils.ToPtr("Europe/London"),
			},
			startPeriod: time.Date(2023, 4, 30, 23, 0, 0, 0, time.UTC),
			endPeriod:   time.Date(2023, 5, 31, 23, 0, 0, 0, time.UTC),
		},
		{
			name:        "monthly Moscow over year",
			currentTime: time.Date(2023, 1, 13, 15, 39, 0, 0, lc),
			limit: domain.Limit{
				Period:   utils.ToPtr(domain.PeriodTypeCALENDARMONTH),
				Timezone: utils.ToPtr("Europe/Moscow"),
			},
			startPeriod: time.Date(2022, 12, 31, 21, 0, 0, 0, time.UTC),
			endPeriod:   time.Date(2023, 1, 31, 21, 0, 0, 0, time.UTC),
		},
		{
			name:        "monthly Moscow over December",
			currentTime: time.Date(2022, 12, 31, 15, 39, 0, 0, lc),
			limit: domain.Limit{
				Period:   utils.ToPtr(domain.PeriodTypeCALENDARMONTH),
				Timezone: utils.ToPtr("Europe/Moscow"),
			},
			startPeriod: time.Date(2022, 11, 30, 21, 0, 0, 0, time.UTC),
			endPeriod:   time.Date(2022, 12, 31, 21, 0, 0, 0, time.UTC),
		},
		{
			name:        "monthly Moscow over January",
			currentTime: time.Date(2023, 1, 31, 15, 39, 0, 0, lc),
			limit: domain.Limit{
				Period:   utils.ToPtr(domain.PeriodTypeCALENDARMONTH),
				Timezone: utils.ToPtr("Europe/Moscow"),
			},
			startPeriod: time.Date(2022, 12, 31, 21, 0, 0, 0, time.UTC),
			endPeriod:   time.Date(2023, 1, 31, 21, 0, 0, 0, time.UTC),
		},
		{
			name:        "monthly Moscow over February",
			currentTime: time.Date(2023, 2, 28, 15, 39, 0, 0, lc),
			limit: domain.Limit{
				Period:   utils.ToPtr(domain.PeriodTypeCALENDARMONTH),
				Timezone: utils.ToPtr("Europe/Moscow"),
			},
			startPeriod: time.Date(2023, 1, 31, 21, 0, 0, 0, time.UTC),
			endPeriod:   time.Date(2023, 2, 28, 21, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			startPeriod, endPeriod, err := generateStartEndPeriods(tt.limit, tt.currentTime)
			if tt.err != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.err)
				return
			}

			require.Equal(t, tt.startPeriod, startPeriod)
			require.Equal(t, tt.endPeriod, endPeriod)
		})
	}
}
