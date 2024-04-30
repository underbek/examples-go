package domain

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/underbek/examples-go/errors"
)

type Attributes map[string]interface{}

func (a Attributes) Value() (driver.Value, error) {
	b, err := json.Marshal(a)
	if err != nil {
		return nil, errors.Wrap(err, errors.TypeInternal, "json.Marshal")
	}

	return b, nil
}

func (a *Attributes) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch data := value.(type) {
	case []byte:
		err := json.Unmarshal(data, &a)
		if err != nil {
			return errors.Wrap(err, errors.TypeInternal, "json.Unmarshal")
		}
	case string:
		err := json.Unmarshal([]byte(data), &a)
		if err != nil {
			return errors.Wrap(err, errors.TypeInternal, "json.Unmarshal")
		}
	default:
		return errors.New(errors.TypeInternal, "type assertion to []byte failed")
	}

	return nil
}
