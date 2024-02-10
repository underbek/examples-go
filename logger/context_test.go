package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDataFromContext(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want map[string]any
	}{
		{
			name: "Nothing found",
			ctx:  context.Background(),
			want: nil,
		},
		{
			name: "Wrong data type",
			ctx:  context.WithValue(context.Background(), contextDataKey, "wrong data"),
			want: nil,
		},
		{
			name: "Something found",
			ctx: context.WithValue(context.Background(), contextDataKey, ContextData{
				fields: map[string]any{
					"field1": 123,
					"field2": true,
					"field0": "value0",
				},
			}),
			want: map[string]any{
				"field1": 123,
				"field2": true,
				"field0": "value0",
			},
		},
		{
			name: "Something found with meta",
			ctx: context.WithValue(context.Background(), contextDataKey, ContextData{
				fields: map[string]any{
					"field1": 123,
				},
				meta: map[string]string{
					"field2": "true",
					"field0": "value0",
				},
			}),
			want: map[string]any{
				"field1": 123,
				Meta: map[string]string{
					"field2": "true",
					"field0": "value0",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, GetDataFromContext(tt.ctx))
		})
	}
}
