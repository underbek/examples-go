package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type Attributes []Attribute

type Attribute struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Scan implements the Scanner interface.
func (a *Attributes) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var result map[string]string
	switch data := value.(type) {
	case []byte:
		if err := json.Unmarshal(data, &result); err != nil {
			return err
		}
	case string:
		if err := json.Unmarshal([]byte(data), &result); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid type")
	}

	for k, v := range result {
		*a = append(*a, Attribute{Name: k, Value: v})
	}

	SortEntities(*a)

	return nil
}

// Value implements the driver Valuer interface.
func (a Attributes) Value() (driver.Value, error) {
	result := make(map[string]string)
	for _, attribute := range a {
		result[attribute.Name] = attribute.Value
	}

	return json.Marshal(result)
}

func SortEntities(entities []Attribute) {
	sort.Slice(entities, func(i, j int) bool {
		return strings.Compare(entities[i].Name, entities[j].Name) == -1
	})
}
