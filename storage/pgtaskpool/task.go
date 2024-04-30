package pgtaskpool

import (
	"encoding/json"
	"fmt"
	"time"
)

/*
ENUM(
default
custom
)
*/
type ScheduleType string

type (
	Trace struct {
		TraceID string            `json:"trace_id"`
		SpanID  string            `json:"span_id"`
		Meta    map[string]string `json:"meta,omitempty"`
	}

	byteTrace struct {
		TraceID [16]byte `json:"trace_id"`
		SpanID  [8]byte  `json:"span_id"`
	}

	Task struct {
		ID                  uint64       `db:"id"`
		TransactionID       uint64       `db:"transaction_id"`
		InitDelaySec        uint         `db:"-"`
		CustomScheduleSlice []uint       `db:"custom_schedule"`
		Attempts            int          `db:"attempts"`
		Type                ScheduleType `db:"schedule_type"`
		Trace               *Trace       `db:"trace_meta"`
		LastErrorCode       *uint64      `db:"last_error_code"`
		LastErrorMessage    *string      `db:"last_error_message"`
		Status              *string      `db:"status"`
		NextProcessingTime  *time.Time   `db:"process_at"`
		LockTime            *time.Time   `db:"lock_time"`
	}
)

func (t *Trace) UnmarshalJSON(data []byte) error {
	var legacy byteTrace
	if err := json.Unmarshal(data, &legacy); err == nil {
		t.TraceID = fmt.Sprintf("%x", legacy.TraceID)
		t.SpanID = fmt.Sprintf("%x", legacy.SpanID)
		return nil
	}

	//to avoid loop in unmarshall
	type alias Trace
	return json.Unmarshal(data, (*alias)(t))
}

func (t *Task) GenerateDefaultSchedule(lifetime uint32) {
	if lifetime == 0 {
		return
	}

	schedule := make([]uint, 0)

	//to prevent the custom schedule from being reset to null, which will subsequently run for two weeks
	if lifetime < fiveSec {
		schedule = append(schedule, uint(lifetime))
		t.CustomScheduleSlice = schedule
		return
	}

	processedSeconds := uint32(0)

	for processedSeconds < lifetime {
		var add uint32
		switch s := processedSeconds; {
		case s < ninetySec:
			add = fiveSec
		case s >= ninetySec && s < fifteenMinutes:
			add = sixtySec
		case s >= fifteenMinutes && s < oneDay:
			add = oneHour
		default:
			add = oneDay
		}

		processedSeconds = processedSeconds + add

		//if the processed seconds is more than lifetime
		//we should make one last call in the end
		//of the lifetime
		if processedSeconds > lifetime {
			//going back to previous processedSeconds value
			processedSeconds = processedSeconds - add

			//calculating final call
			finalTime := uint(lifetime - processedSeconds)
			if finalTime > 0 {
				schedule = append(schedule, finalTime)
			}

			break
		}

		schedule = append(schedule, uint(add))
	}

	t.CustomScheduleSlice = schedule
}
