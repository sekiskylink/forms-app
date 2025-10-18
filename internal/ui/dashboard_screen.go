package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"forms-app/internal/forms"
)

// DashboardScreen lists available forms and adds an offline drafts uploader.
func DashboardScreen(a fyne.App, formDefs map[string][]forms.Section, banner fyne.CanvasObject, openForm func(name string)) fyne.CanvasObject {
	formNames := make([]string, 0, len(formDefs))
	for name := range formDefs {
		formNames = append(formNames, name)
	}

	var buttons []fyne.CanvasObject
	for _, name := range formNames {
		n := name // capture
		btn := widget.NewButton("📝 "+n, func() { openForm(n) })
		buttons = append(buttons, btn)
	}

	if len(buttons) == 0 {
		buttons = append(buttons, widget.NewLabel("⚠️ No forms available."))
	}

	scroll := container.NewVScroll(container.NewVBox(buttons...))
	scroll.SetMinSize(fyne.NewSize(300, 400))

	header := widget.NewLabelWithStyle("Available Forms", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// 📦 Draft button (only if drafts exist)
	apiURL := "https://example.com/api/forms/submit" // ← your real endpoint
	drafts, _ := forms.LoadDrafts(a)
	var draftBtn fyne.CanvasObject
	if len(drafts) > 0 {
		count := len(drafts)
		draftBtn = widget.NewButton(fmt.Sprintf("📂 View %d Draft%s", count, plural(count)), func() {
			screen := DraftsScreen(a, apiURL, a.Driver().AllWindows()[0], func() {
				// Return to dashboard when user presses Back
				main := DashboardScreen(a, formDefs, banner, openForm)
				a.Driver().AllWindows()[0].SetContent(main)
			})
			a.Driver().AllWindows()[0].SetContent(screen)
		})
	} else {
		draftBtn = widget.NewLabel("No pending drafts")
	}

	mainContent := container.NewVBox(
		header,
		draftBtn,
		container.NewPadded(scroll),
	)

	// Banner on top if available
	if banner != nil {
		return container.NewBorder(
			banner, // top
			nil,    // bottom
			nil,    // left
			nil,    // right
			mainContent,
		)
	}

	return mainContent
}

// plural adds "s" for plural count
func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
