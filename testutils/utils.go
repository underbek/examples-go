package testutils

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func JSONEq(t *testing.T, expected, actual any) bool {
	return assert.JSONEq(t, jsonMarshal(t, expected), jsonMarshal(t, actual))
}

func jsonMarshal(t *testing.T, data any) string {
	switch v := data.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case io.Reader:
		body, err := io.ReadAll(v)
		if err != nil {
			t.Fatal(err)
		}
		return string(body)
	default:
		res, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return string(res)
	}
}
