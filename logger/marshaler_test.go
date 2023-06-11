package logger

import (
	"math"
	"testing"
)

func TestSliceStringer_String(t *testing.T) {
	tests := []struct {
		name string
		am   SliceStringer
		want string
	}{
		{
			"Nil data",
			SliceStringer(nil),
			"null",
		},
		{
			"Empty data",
			SliceStringer([]interface{}{}),
			"[]",
		},
		{
			"Various data",
			SliceStringer([]interface{}{
				123,
				"Apple",
				[]int{1, 2, 3},
				true,
				nil,
				map[string]interface{}{
					"orange": 456,
				},
			}),
			`[123,"Apple",[1,2,3],true,null,{"orange":456}]`,
		},
		{
			"Error when invalid value",
			SliceStringer([]interface{}{
				math.Inf(1),
			}),
			"marshaling error: json: unsupported value: +Inf",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.am.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
