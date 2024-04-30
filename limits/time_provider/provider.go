package time_provider

import "time"

//go:generate mockery --name TimeProvider --structname TimeProviderMock --filename time_provider_mock.go --inpackage
type TimeProvider interface {
	Now() time.Time
}

type RealTime struct{}

func (RealTime) Now() time.Time {
	return time.Now()
}
