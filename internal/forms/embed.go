package forms

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed assets/forms.json
var embeddedForms []byte

func LoadFromEmbedded() (map[string][]Section, error) {
	var bundle FormBundle
	if err := json.Unmarshal(embeddedForms, &bundle); err == nil && bundle.Forms != nil {
		return bundle.Forms, nil
	}

	var legacy map[string][]Section
	if err := json.Unmarshal(embeddedForms, &legacy); err == nil {
		return legacy, nil
	}
	return nil, fmt.Errorf("embedded forms are invalid")
}
