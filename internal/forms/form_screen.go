package forms

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func makeAdaptiveGrid(w fyne.Window, items []fyne.CanvasObject) fyne.CanvasObject {
	width := w.Canvas().Size().Width
	var cols int
	switch {
	case width > 900:
		cols = 3
	case width > 500:
		cols = 2
	default:
		cols = 1
	}
	return container.NewGridWithColumns(cols, items...)
}

// BuildForm builds a form with section tabs.
// Supports "grid"/"stack" layouts, responsive columns, and auto-hides tabs if only one section.
func BuildForm(a fyne.App, formName string, sections []Section, onSubmit func(data map[string]string)) fyne.CanvasObject {
	allText := make(map[string]*widget.Entry)
	allSelect := make(map[string]*widget.Select)
	allDate := make(map[string]*widget.DateEntry)
	allBool := make(map[string]*widget.Check)

	buildSectionContent := func(sec Section) fyne.CanvasObject {
		var items []fyne.CanvasObject

		for _, f := range sec.Fields {
			var field fyne.CanvasObject

			switch f.Type {
			case "text", "number", "multiline":
				e := widget.NewEntry()
				if f.Type == "multiline" {
					e.MultiLine = true
				}
				if f.Type == "number" {
					e.SetPlaceHolder("Enter number...")
					e.OnChanged = func(text string) {
						clean := strings.Map(func(r rune) rune {
							if (r >= '0' && r <= '9') || r == '.' {
								return r
							}
							return -1
						}, text)
						if clean != text {
							e.SetText(clean)
						}
					}
				}
				field = container.NewVBox(widget.NewLabel(f.Label), e)
				allText[f.Label] = e

			case "select":
				s := widget.NewSelect(f.Options, func(string) {})
				s.PlaceHolder = "Select..."
				field = container.NewVBox(widget.NewLabel(f.Label), s)
				allSelect[f.Label] = s

			case "date":
				d := widget.NewDateEntry()
				field = container.NewVBox(widget.NewLabel(f.Label), d)
				allDate[f.Label] = d

			case "boolean":
				c := widget.NewCheck(f.Label, func(bool) {})
				field = c
				allBool[f.Label] = c
			}

			items = append(items, field)
		}

		layout := strings.ToLower(sec.Layout)
		switch layout {
		case "grid":
			// Use adaptive or fixed columns
			if sec.Columns > 0 {
				return container.NewGridWithColumns(sec.Columns, items...)
			}
			// Fallback to our manual adaptive grid
			return makeAdaptiveGrid(a.Driver().AllWindows()[0], items) // responsive 1–3 cols
		default:
			return container.NewVBox(items...)
		}
	}

	var formContent fyne.CanvasObject

	if len(sections) == 1 {
		formContent = buildSectionContent(sections[0])
	} else {
		tabs := container.NewAppTabs()
		for _, sec := range sections {
			tab := container.NewTabItem(sec.Title, buildSectionContent(sec))
			tabs.Append(tab)
		}
		formContent = tabs
	}

	submit := widget.NewButton("Submit", func() {
		data := map[string]string{}
		for k, v := range allText {
			data[k] = v.Text
		}
		for k, v := range allSelect {
			data[k] = v.Selected
		}
		for k, v := range allDate {
			if v.Date != nil {
				data[k] = v.Date.Format("2006-01-02")
			}
		}
		for k, v := range allBool {
			data[k] = strconv.FormatBool(v.Checked)
		}

		// Gather fields for validation
		var allFields []Field
		for _, sec := range sections {
			allFields = append(allFields, sec.Fields...)
		}

		if err := validateForm(allFields, allText, allSelect, allDate, allBool); err != nil {
			dialog.ShowInformation("Validation Error", err.Error(), a.Driver().AllWindows()[0])
			return
		}

		onSubmit(data)
	})

	return container.NewBorder(
		nil,
		container.NewVBox(submit),
		nil,
		nil,
		container.NewVBox(
			widget.NewLabelWithStyle(formName+" Form", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			formContent,
		),
	)
}

// validateForm applies validation rules to all field types.
func validateForm(
	fields []Field,
	textEntries map[string]*widget.Entry,
	selectEntries map[string]*widget.Select,
	dateEntries map[string]*widget.DateEntry,
	boolEntries map[string]*widget.Check,
) error {
	for _, f := range fields {
		v := f.Validation
		var val string

		// pick correct value
		if e, ok := textEntries[f.Label]; ok {
			val = e.Text
		} else if s, ok := selectEntries[f.Label]; ok {
			val = s.Selected
		} else if d, ok := dateEntries[f.Label]; ok && d.Date != nil {
			val = d.Date.Format("2006-01-02")
		} else if c, ok := boolEntries[f.Label]; ok {
			val = strconv.FormatBool(c.Checked)
		}

		// Required validation
		if v.Required {
			switch f.Type {
			case "boolean":
				if val == "false" {
					return fmt.Errorf("'%s' must be checked to proceed", f.Label)
				}
			default:
				if val == "" {
					return fmt.Errorf("'%s' is required", f.Label)
				}
			}
		}

		// Text length
		if v.MinLength > 0 && len(val) < v.MinLength {
			return fmt.Errorf("'%s' must be at least %d characters", f.Label, v.MinLength)
		}
		if v.MaxLength > 0 && len(val) > v.MaxLength {
			return fmt.Errorf("'%s' must be no more than %d characters", f.Label, v.MaxLength)
		}

		// Numeric
		if f.Type == "number" && val != "" {
			num, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return fmt.Errorf("'%s' must be numeric", f.Label)
			}
			if v.Min != 0 && num < v.Min {
				return fmt.Errorf("'%s' must be ≥ %.2f", f.Label, v.Min)
			}
			if v.Max != 0 && num > v.Max {
				return fmt.Errorf("'%s' must be ≤ %.2f", f.Label, v.Max)
			}
		}

		// Regex pattern
		if v.Pattern != "" && val != "" {
			re, err := regexp.Compile(v.Pattern)
			if err == nil && !re.MatchString(val) {
				return fmt.Errorf("'%s' does not match expected format", f.Label)
			}
		}

		// Date validation
		if f.Type == "date" && val != "" {
			dateVal, err := time.Parse("2006-01-02", val)
			if err != nil {
				return fmt.Errorf("'%s' is not a valid date", f.Label)
			}
			if v.MinDate != "" {
				minDate, _ := time.Parse("2006-01-02", v.MinDate)
				if dateVal.Before(minDate) {
					return fmt.Errorf("'%s' must be after %s", f.Label, v.MinDate)
				}
			}
			if v.MaxDate != "" {
				maxDate, _ := time.Parse("2006-01-02", v.MaxDate)
				if dateVal.After(maxDate) {
					return fmt.Errorf("'%s' must be before %s", f.Label, v.MaxDate)
				}
			}
		}
	}
	return nil
}
