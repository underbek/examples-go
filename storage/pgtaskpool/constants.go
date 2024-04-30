package pgtaskpool

import (
	"errors"
	"time"
)

const (
	fiveSec   = 5
	sixtySec  = 60
	ninetySec = 90

	fifteenMinutes  = sixtySec * 15
	oneHour         = sixtySec * 60
	oneDay          = oneHour * 24
	defaultLifetime = uint32(oneHour * 24 * 14) //2 weeks

	rollbackTimeout = fiveSec * time.Second

	statusFailed    = "failed"
	statusSuccess   = "success"
	statusProcess   = "process"
	statusCancelled = "canceled"

	defaultTaskTable = "tasks"
)

var ErrNeedJustToRetry = errors.New("need just to retry")
