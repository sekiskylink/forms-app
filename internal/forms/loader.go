package forms

type Validation struct {
	Required         bool    `json:"required"`
	MinLength        int     `json:"minLength"`
	MaxLength        int     `json:"maxLength"`
	Min              float64 `json:"min"`
	Max              float64 `json:"max"`
	Pattern          string  `json:"pattern"`
	MinDate          string  `json:"minDate"`
	MaxDate          string  `json:"maxDate"`
	GreaterThanField string  `json:"greaterThanField"` // A > B
	LessThanField    string  `json:"lessThanField"`    // A < B
	BeforeField      string  `json:"beforeField"`      // A before B (dates)
	AfterField       string  `json:"afterField"`
	EqualToField     string  `json:"equalToField"`    // equal
	NotEqualToField  string  `json:"notEqualToField"` // not equal
	Formula          string  `json:"formula"`         // e.g., "A + B == C"
	ErrorMessage     string  `json:"errorMessage"`    // optional custom message
}

type Field struct {
	ID         string     `json:"id"`
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

type FormDefinition struct {
	Meta     FormMeta  `json:"meta"`
	Sections []Section `json:"sections"`
}

type FormMeta struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type FormBundle struct {
	Version     string                    `json:"version"`
	LastUpdated string                    `json:"lastUpdated,omitempty"`
	Forms       map[string]FormDefinition `json:"forms"`
	FormOrder   []string                  `json:"form_order,omitempty"`
}
