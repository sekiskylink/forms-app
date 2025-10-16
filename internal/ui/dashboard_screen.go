package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"forms-app/internal/forms"
)

// DashboardScreen lists available forms
func DashboardScreen(a fyne.App, formDefs map[string][]forms.Section, openForm func(name string)) fyne.CanvasObject {
	formNames := make([]string, 0, len(formDefs))
	for name := range formDefs {
		formNames = append(formNames, name)
	}

	var buttons []fyne.CanvasObject
	for _, name := range formNames {
		n := name // capture loop variable
		btn := widget.NewButton("📝 "+n, func() { openForm(n) })
		buttons = append(buttons, btn)
	}

	listContainer := container.NewVBox(buttons...)
	scroll := container.NewVScroll(listContainer)
	scroll.SetMinSize(fyne.NewSize(300, 400))

	return container.NewBorder(
		widget.NewLabelWithStyle("Available Forms", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		nil, nil, nil,
		container.NewPadded(scroll),
	)
}
