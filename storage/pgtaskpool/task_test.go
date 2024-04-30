package pgtaskpool

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrace_UnmarshalTrace(t *testing.T) {
	tests := []struct {
		name,
		value,
		expectedTraceID,
		expectedSpanID string
		wantErr bool
	}{
		{
			name:            "Success string",
			value:           `{"span_id": "9d2796021d8f4b68", "trace_id": "f762a44e67302a556a1a9fc753a4102d"}`,
			expectedTraceID: "f762a44e67302a556a1a9fc753a4102d",
			expectedSpanID:  "9d2796021d8f4b68",
		},
		{
			name:            "Success byte",
			value:           `{"span_id": [16, 246, 87, 226, 63, 249, 111, 120], "trace_id": [74, 206, 227, 186, 225, 89, 45, 5, 153, 25, 206, 211, 107, 63, 12, 72]}`,
			expectedTraceID: "4acee3bae1592d059919ced36b3f0c48",
			expectedSpanID:  "10f657e23ff96f78",
		},
		{
			name:            "Cropped byte",
			value:           `{"span_id": [16, 246, 87, 226, 63, 249], "trace_id": [74, 206, 227]}`,
			expectedTraceID: "4acee300000000000000000000000000",
			expectedSpanID:  "10f657e23ff90000",
		},
		{
			name:    "fail",
			value:   `{"span_id": {"span_sub": 123}, "trace_id": [74, 206, 227]}`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var trace Trace
			data := []byte(tt.value)

			err := json.Unmarshal(data, &trace)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedTraceID, trace.TraceID)
			assert.Equal(t, tt.expectedSpanID, trace.SpanID)
		})
	}
}

func TestTask_GenerateDefaultSchedule(t *testing.T) {
	tests := []struct {
		name         string
		lifetime     uint32
		wantSchedule []uint
	}{
		{
			name:         "60 seconds",
			lifetime:     uint32(time.Minute.Seconds()),
			wantSchedule: []uint{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5},
		},
		{
			name:     "4 minutes",
			lifetime: uint32(time.Minute.Seconds()) * 4,
			wantSchedule: []uint{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
				5, 5, 5, 5, 5, 60, 60, 30},
		},
		{
			name:     "30 minutes",
			lifetime: uint32(time.Minute.Seconds()) * 30,

			wantSchedule: []uint{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
				5, 5, 5, 5, 5, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60,
				60, 60, 60, 60, 870,
			},
		},
		{
			name:     "1 hour",
			lifetime: uint32(time.Minute.Seconds()) * 60,
			wantSchedule: []uint{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
				5, 5, 5, 5, 5, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60,
				60, 60, 60, 2670,
			},
		},
		{
			name:     "5 hours",
			lifetime: uint32(time.Hour.Seconds()) * 5,

			wantSchedule: []uint{5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
				5, 5, 5, 5, 5, 60, 60, 60, 60, 60, 60, 60, 60, 60, 60,
				60, 60, 60, 60, 3600, 3600, 3600, 3600, 2670,
			},
		},
		{
			name:         "less than 5 secs",
			lifetime:     1,
			wantSchedule: []uint{1},
		},
		{
			name:         "5 secs",
			lifetime:     5,
			wantSchedule: []uint{5},
		},
		{
			name:         "7 secs",
			lifetime:     7,
			wantSchedule: []uint{5, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Task{
				TransactionID:       123,
				InitDelaySec:        0,
				CustomScheduleSlice: []uint{},
			}
			tr.GenerateDefaultSchedule(tt.lifetime)
			require.Equal(t, tt.wantSchedule, tr.CustomScheduleSlice)
		})
	}
}
