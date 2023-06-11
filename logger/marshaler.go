package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type SliceStringer []interface{}

func (ss SliceStringer) String() string {
	b := new(bytes.Buffer)
	e := json.NewEncoder(b)
	e.SetEscapeHTML(false)

	if err := e.Encode(ss); err != nil {
		return fmt.Sprint("marshaling error: ", err.Error())
	}
	return strings.TrimRight(b.String(), "\n")
}
