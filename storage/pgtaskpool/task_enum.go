// Code generated by go-enum DO NOT EDIT.
// Version:
// Revision:
// Build Date:
// Built By:

package pgtaskpool

import (
	"database/sql/driver"
	"errors"
	"fmt"
)

const (
	// ScheduleTypeDefault is a ScheduleType of type default.
	ScheduleTypeDefault ScheduleType = "default"
	// ScheduleTypeCustom is a ScheduleType of type custom.
	ScheduleTypeCustom ScheduleType = "custom"
)

var ErrInvalidScheduleType = errors.New("not a valid ScheduleType")

// String implements the Stringer interface.
func (x ScheduleType) String() string {
	return string(x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x ScheduleType) IsValid() bool {
	_, err := ParseScheduleType(string(x))
	return err == nil
}

var _ScheduleTypeValue = map[string]ScheduleType{
	"default": ScheduleTypeDefault,
	"custom":  ScheduleTypeCustom,
}

// ParseScheduleType attempts to convert a string to a ScheduleType.
func ParseScheduleType(name string) (ScheduleType, error) {
	if x, ok := _ScheduleTypeValue[name]; ok {
		return x, nil
	}
	return ScheduleType(""), fmt.Errorf("%s is %w", name, ErrInvalidScheduleType)
}

// MarshalText implements the text marshaller method.
func (x ScheduleType) MarshalText() ([]byte, error) {
	return []byte(string(x)), nil
}

// UnmarshalText implements the text unmarshaller method.
func (x *ScheduleType) UnmarshalText(text []byte) error {
	tmp, err := ParseScheduleType(string(text))
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}

var errScheduleTypeNilPtr = errors.New("value pointer is nil") // one per type for package clashes

// Scan implements the Scanner interface.
func (x *ScheduleType) Scan(value interface{}) (err error) {
	if value == nil {
		*x = ScheduleType("")
		return
	}

	// A wider range of scannable types.
	// driver.Value values at the top of the list for expediency
	switch v := value.(type) {
	case string:
		*x, err = ParseScheduleType(v)
	case []byte:
		*x, err = ParseScheduleType(string(v))
	case ScheduleType:
		*x = v
	case *ScheduleType:
		if v == nil {
			return errScheduleTypeNilPtr
		}
		*x = *v
	case *string:
		if v == nil {
			return errScheduleTypeNilPtr
		}
		*x, err = ParseScheduleType(*v)
	default:
		return errors.New("invalid type for ScheduleType")
	}

	return
}

// Value implements the driver Valuer interface.
func (x ScheduleType) Value() (driver.Value, error) {
	return x.String(), nil
}
