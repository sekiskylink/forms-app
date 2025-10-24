package forms

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed assets/forms.json
var embeddedForms []byte

func LoadFromEmbedded() (map[string]FormDefinition, []string, error) {
	var bundle FormBundle
	if err := json.Unmarshal(embeddedForms, &bundle); err == nil && bundle.Forms != nil {
		return bundle.Forms, bundle.FormOrder, nil
	}

	return nil, nil, fmt.Errorf("embedded forms are invalid")
}
