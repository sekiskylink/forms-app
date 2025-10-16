package forms

import (
	"encoding/json"
	"os"
)

type Validation struct {
	Required  bool    `json:"required"`
	MinLength int     `json:"minLength"`
	MaxLength int     `json:"maxLength"`
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
	Pattern   string  `json:"pattern"`
	MinDate   string  `json:"minDate"`
	MaxDate   string  `json:"maxDate"`
}

type Field struct {
	Label      string     `json:"label"`
	Type       string     `json:"type"`
	Options    []string   `json:"options"`
	Validation Validation `json:"validation"`
}

type Section struct {
	Title   string  `json:"title"`
	Layout  string  `json:"layout"`  // "stack" or "grid"
	Columns int     `json:"columns"` // optional, for grid layout
	Fields  []Field `json:"fields"`
}

// Each form now maps to a list of sections
func LoadForms(path string) (map[string][]Section, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var result map[string][]Section
	err = json.Unmarshal(b, &result)
	return result, err
}
