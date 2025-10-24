package forms

import (
	"fmt"
	"image/color"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/Knetic/govaluate"
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
func BuildForm(
	a fyne.App,
	formName string,
	sections []Section,
	onSubmit func(data map[string]string),
	prefill ...map[string]string, // optional prefill values
) fyne.CanvasObject {
	var values map[string]string
	if len(prefill) > 0 {
		values = prefill[0]
	}

	allText := make(map[string]*widget.Entry)
	allSelect := make(map[string]*widget.Select)
	allDate := make(map[string]*widget.DateEntry)
	allBool := make(map[string]*widget.Check)
	errorLabels := make(map[string]*widget.Label)
	overlayRects := make(map[string]*canvas.Rectangle)

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
				if f.Validation.MaxLength > 0 {
					maxLen := f.Validation.MaxLength
					oldHandler := e.OnChanged
					e.OnChanged = func(text string) {
						if oldHandler != nil {
							oldHandler(text)
						}
						if len(e.Text) > maxLen {
							e.SetText(e.Text[:maxLen])
						}
					}
				}
				// prefill value
				if val, ok := values[f.ID]; ok {
					e.SetText(val)
				}

				errLbl := widget.NewLabel("")
				errLbl.TextStyle = fyne.TextStyle{Italic: true}
				errLbl.Hide()
				errorLabels[f.ID] = errLbl

				content := container.NewVBox(widget.NewLabel(f.Label), e, errLbl)
				overlay := canvas.NewRectangle(color.NRGBA{255, 0, 0, 40})
				overlay.Hide()
				overlayRects[f.ID] = overlay
				field = container.NewStack(content, overlay)
				allText[f.ID] = e

			case "select":
				s := widget.NewSelect(f.Options, func(string) {})
				s.PlaceHolder = "Select..."
				if val, ok := values[f.ID]; ok {
					s.SetSelected(val)
				}
				errLbl := widget.NewLabel("")
				errLbl.TextStyle = fyne.TextStyle{Italic: true}
				errLbl.Hide()
				errorLabels[f.ID] = errLbl
				content := container.NewVBox(widget.NewLabel(f.Label), s, errLbl)
				overlay := canvas.NewRectangle(color.NRGBA{255, 0, 0, 40})
				overlay.Hide()
				overlayRects[f.ID] = overlay
				field = container.NewStack(content, overlay)
				allSelect[f.ID] = s

			case "date":
				d := widget.NewDateEntry()
				if val, ok := values[f.ID]; ok && val != "" {
					if parsed, err := time.Parse("2006-01-02", val); err == nil {
						d.SetDate(&parsed)
					}
				}
				errLbl := widget.NewLabel("")
				errLbl.TextStyle = fyne.TextStyle{Italic: true}
				errLbl.Hide()
				errorLabels[f.ID] = errLbl
				content := container.NewVBox(widget.NewLabel(f.Label), d, errLbl)
				overlay := canvas.NewRectangle(color.NRGBA{255, 0, 0, 40})
				overlay.Hide()
				overlayRects[f.ID] = overlay
				field = container.NewStack(content, overlay)
				allDate[f.ID] = d

			case "boolean":
				c := widget.NewCheck(f.Label, func(bool) {})
				if val, ok := values[f.ID]; ok {
					c.SetChecked(val == "true" || val == "1" || strings.EqualFold(val, "yes"))
				}
				errLbl := widget.NewLabel("")
				errLbl.TextStyle = fyne.TextStyle{Italic: true}
				errLbl.Hide()
				errorLabels[f.ID] = errLbl
				content := container.NewVBox(c, errLbl)
				overlay := canvas.NewRectangle(color.NRGBA{255, 0, 0, 40})
				overlay.Hide()
				overlayRects[f.ID] = overlay
				field = container.NewStack(content, overlay)
				allBool[f.ID] = c
			}
			items = append(items, field)
		}

		if strings.ToLower(sec.Layout) == "grid" {
			if sec.Columns > 0 {
				return container.NewGridWithColumns(sec.Columns, items...)
			}
			return makeAdaptiveGrid(a.Driver().AllWindows()[0], items)
		}
		return container.NewVBox(items...)
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
		for id, lbl := range errorLabels {
			lbl.Hide()
			if r, ok := overlayRects[id]; ok {
				r.Hide()
				canvas.Refresh(r)
			}
		}

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

		var allFields []Field
		for _, sec := range sections {
			allFields = append(allFields, sec.Fields...)
		}

		fieldErrs, err := validateForm(allFields, allText, allSelect, allDate, allBool)
		if err != nil && len(fieldErrs) > 0 {
			for id, msg := range fieldErrs {
				if lbl, ok := errorLabels[id]; ok {
					lbl.SetText("⚠ " + msg)
					lbl.Show()
				}
				if r, ok := overlayRects[id]; ok {
					r.Show()
					canvas.Refresh(r)
				}
			}
			dialog.ShowError(fmt.Errorf("Please correct the highlighted fields."), a.Driver().AllWindows()[0])
			return
		}

		apiURL := "https://example.com/api/forms/submit"
		go func() {
			err := SubmitForm(a, apiURL, formName, data)
			fyne.Do(func() {
				if err != nil {
					if strings.Contains(err.Error(), "offline mode") {
						dialog.ShowInformation("📥 Saved Offline",
							"No network — form stored locally for later upload.",
							a.Driver().AllWindows()[0])
					} else {
						dialog.ShowError(fmt.Errorf("Submission failed: %v", err),
							a.Driver().AllWindows()[0])
					}
					return
				}
				dialog.ShowInformation("✅ Success",
					"Form submitted successfully!",
					a.Driver().AllWindows()[0])
				onSubmit(data)
			})
		}()
	})

	saveBtn := widget.NewButton("💾 Save Draft", func() {
		payload := map[string]any{
			"form":      formName,
			"data":      collectData(allText, allSelect, allDate, allBool),
			"timestamp": time.Now().Format(time.RFC3339),
		}
		if err := SaveTaggedDraft(a, formName, payload); err != nil {
			dialog.ShowError(fmt.Errorf("Failed to save draft: %v", err), a.Driver().AllWindows()[0])
		} else {
			dialog.ShowInformation("💾 Saved", "Draft saved locally for later submission.", a.Driver().AllWindows()[0])
		}
	})

	buttons := container.NewGridWithColumns(2, submit, saveBtn)

	return container.NewBorder(
		nil,
		container.NewVBox(buttons),
		nil, nil,
		container.NewVBox(
			widget.NewLabelWithStyle(formName+" Form", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			formContent,
		),
	)
}

// collectData consolidates form values.
func collectData(
	allText map[string]*widget.Entry,
	allSelect map[string]*widget.Select,
	allDate map[string]*widget.DateEntry,
	allBool map[string]*widget.Check,
) map[string]string {
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
	return data
}

// validateForm applies validation rules to all field types.
func validateForm(
	fields []Field,
	textEntries map[string]*widget.Entry,
	selectEntries map[string]*widget.Select,
	dateEntries map[string]*widget.DateEntry,
	boolEntries map[string]*widget.Check,
) (map[string]string, error) {
	fieldErrors := make(map[string]string)
	values := make(map[string]string)

	// Build a map of all values for cross-field access
	for _, f := range fields {
		if e, ok := textEntries[f.ID]; ok {
			values[f.ID] = e.Text
		} else if s, ok := selectEntries[f.ID]; ok {
			values[f.ID] = s.Selected
		} else if d, ok := dateEntries[f.ID]; ok && d.Date != nil {
			values[f.ID] = d.Date.Format("2006-01-02")
		} else if c, ok := boolEntries[f.ID]; ok {
			values[f.ID] = strconv.FormatBool(c.Checked)
		}
	}

	for _, f := range fields {
		v := f.Validation
		val := values[f.ID]

		// ---------- Required ----------
		if v.Required {
			switch f.Type {
			case "boolean":
				if val == "false" {
					fieldErrors[f.ID] = fmt.Sprintf("'%s' must be checked to proceed", f.Label)
				}
			default:
				if val == "" {
					fieldErrors[f.ID] = fmt.Sprintf("'%s' is required", f.Label)
				}
			}
		}

		// ---------- Length validation ----------
		if v.MinLength > 0 && len(val) < v.MinLength {
			fieldErrors[f.ID] = fmt.Sprintf("'%s' must be at least %d characters", f.Label, v.MinLength)
		}
		if v.MaxLength > 0 && len(val) > v.MaxLength {
			fieldErrors[f.ID] = fmt.Sprintf("'%s' must be no more than %d characters", f.Label, v.MaxLength)
		}

		// ---------- Numeric range ----------
		if f.Type == "number" && val != "" {
			num, err := strconv.ParseFloat(val, 64)
			if err != nil {
				fieldErrors[f.ID] = fmt.Sprintf("'%s' must be numeric", f.Label)
			} else {
				if v.Min != 0 && num < v.Min {
					fieldErrors[f.ID] = fmt.Sprintf("'%s' must be ≥ %.2f", f.Label, v.Min)
				}
				if v.Max != 0 && num > v.Max {
					fieldErrors[f.ID] = fmt.Sprintf("'%s' must be ≤ %.2f", f.Label, v.Max)
				}
			}
		}

		// ---------- Regex ----------
		if v.Pattern != "" && val != "" {
			re, err := regexp.Compile(v.Pattern)
			if err == nil && !re.MatchString(val) {
				fieldErrors[f.ID] = fmt.Sprintf("'%s' does not match expected format", f.Label)
			}
		}

		// ---------- Date range ----------
		if f.Type == "date" && val != "" {
			dateVal, err := time.Parse("2006-01-02", val)
			if err != nil {
				fieldErrors[f.ID] = fmt.Sprintf("'%s' is not a valid date", f.Label)
			} else {
				if v.MinDate != "" {
					minDate, _ := time.Parse("2006-01-02", v.MinDate)
					if dateVal.Before(minDate) {
						fieldErrors[f.ID] = fmt.Sprintf("'%s' must be after %s", f.Label, v.MinDate)
					}
				}
				if v.MaxDate != "" {
					maxDate, _ := time.Parse("2006-01-02", v.MaxDate)
					if dateVal.After(maxDate) {
						fieldErrors[f.ID] = fmt.Sprintf("'%s' must be before %s", f.Label, v.MaxDate)
					}
				}
			}
		}

		// ---------- Cross-field numeric/date checks ----------
		if v.GreaterThanField != "" {
			otherVal, ok := values[v.GreaterThanField]
			if ok && otherVal != "" && val != "" {
				if f.Type == "number" {
					numA, _ := strconv.ParseFloat(val, 64)
					numB, _ := strconv.ParseFloat(otherVal, 64)
					if numA <= numB {
						fieldErrors[f.ID] = fmt.Sprintf("'%s' must be greater than '%s'", f.Label, v.GreaterThanField)
					}
				} else if f.Type == "date" {
					dateA, _ := time.Parse("2006-01-02", val)
					dateB, _ := time.Parse("2006-01-02", otherVal)
					if !dateA.After(dateB) {
						fieldErrors[f.ID] = fmt.Sprintf("'%s' must be after '%s'", f.Label, v.GreaterThanField)
					}
				}
			}
		}

		if v.LessThanField != "" {
			otherVal, ok := values[v.LessThanField]
			if ok && otherVal != "" && val != "" {
				if f.Type == "number" {
					numA, _ := strconv.ParseFloat(val, 64)
					numB, _ := strconv.ParseFloat(otherVal, 64)
					if numA >= numB {
						fieldErrors[f.ID] = fmt.Sprintf("'%s' must be less than '%s'", f.Label, v.LessThanField)
					}
				} else if f.Type == "date" {
					dateA, _ := time.Parse("2006-01-02", val)
					dateB, _ := time.Parse("2006-01-02", otherVal)
					if !dateA.Before(dateB) {
						fieldErrors[f.ID] = fmt.Sprintf("'%s' must be before '%s'", f.Label, v.LessThanField)
					}
				}
			}
		}

		// ---------- Date-specific cross checks ----------
		if v.BeforeField != "" {
			otherVal, ok := values[v.BeforeField]
			if ok && otherVal != "" && val != "" {
				dateA, err1 := time.Parse("2006-01-02", val)
				dateB, err2 := time.Parse("2006-01-02", otherVal)
				if err1 == nil && err2 == nil && !dateA.Before(dateB) {
					fieldErrors[f.ID] = fmt.Sprintf("'%s' must be before '%s'", f.Label, v.BeforeField)
				}
			}
		}

		if v.AfterField != "" {
			otherVal, ok := values[v.AfterField]
			if ok && otherVal != "" && val != "" {
				dateA, err1 := time.Parse("2006-01-02", val)
				dateB, err2 := time.Parse("2006-01-02", otherVal)
				if err1 == nil && err2 == nil && !dateA.After(dateB) {
					fieldErrors[f.ID] = fmt.Sprintf("'%s' must be after '%s'", f.Label, v.AfterField)
				}
			}
		}

		// ---------- Equal / Not Equal ----------
		if v.EqualToField != "" {
			otherVal := values[v.EqualToField]
			if val != "" && otherVal != "" && val != otherVal {
				fieldErrors[f.ID] = fmt.Sprintf("'%s' must equal '%s'", f.Label, v.EqualToField)
			}
		}

		if v.NotEqualToField != "" {
			otherVal := values[v.NotEqualToField]
			if val != "" && otherVal != "" && val == otherVal {
				fieldErrors[f.ID] = fmt.Sprintf("'%s' must not equal '%s'", f.Label, v.NotEqualToField)
			}
		}

		// ---------- Formula-based validation ----------
		if v.Formula != "" {
			ok, err := evalFormula(v.Formula, values)
			if err != nil {
				fieldErrors[f.ID] = fmt.Sprintf("Invalid formula for '%s': %v", f.Label, err)
			} else if !ok {
				if v.ErrorMessage != "" {
					fieldErrors[f.ID] = v.ErrorMessage
				} else {
					fieldErrors[f.ID] = fmt.Sprintf("Formula validation failed for '%s'", f.Label)
				}
			}
		}
	}

	if len(fieldErrors) > 0 {
		return fieldErrors, fmt.Errorf("one or more fields are invalid")
	}

	return nil, nil
}

// evalFormula safely evaluates a logical or arithmetic formula and returns whether it’s true.
func evalFormula(expr string, values map[string]string) (bool, error) {
	// Parse the formula expression
	expression, err := govaluate.NewEvaluableExpression(expr)
	if err != nil {
		return false, fmt.Errorf("invalid formula syntax: %w", err)
	}

	// Build parameters for all known field values
	parameters := make(map[string]interface{})
	for id, value := range values {
		if value == "" {
			parameters[id] = 0.0 // treat empty as zero
			continue
		}

		// Try number
		if num, err := strconv.ParseFloat(value, 64); err == nil {
			parameters[id] = num
			continue
		}

		// Try boolean
		if value == "true" || value == "false" {
			parameters[id] = (value == "true")
			continue
		}

		// Try date (parse as timestamp)
		if t, err := time.Parse("2006-01-02", value); err == nil {
			parameters[id] = float64(t.Unix())
			continue
		}

		// Fallback to string
		parameters[id] = value
	}

	// Evaluate safely
	result, err := expression.Evaluate(parameters)
	if err != nil {
		return false, fmt.Errorf("formula evaluation error: %w", err)
	}

	switch val := result.(type) {
	case bool:
		return val, nil
	case float64:
		return val != 0, nil
	default:
		return false, fmt.Errorf("unexpected formula result type: %T", val)
	}
}
