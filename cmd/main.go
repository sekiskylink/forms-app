package main

import (
	"fmt"
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"forms-app/internal/forms"
	"forms-app/internal/ui"
)

// statusBanner shows a horizontal, visible, dismissible banner at the top.
func statusBanner(source string, w fyne.Window) fyne.CanvasObject {
	var msg string
	var bg color.Color

	switch source {
	case "api":
		msg = "🟢 Online Mode – Loaded from API"
		bg = color.NRGBA{0, 200, 0, 80}
	case "cache":
		msg = "🟡 Offline Mode – Loaded from Cache"
		bg = color.NRGBA{255, 210, 0, 80}
	case "embedded":
		msg = "🔴 Offline Mode – Using Embedded Forms"
		bg = color.NRGBA{255, 0, 0, 80}
	default:
		msg = "⚠️ Unable to load forms"
		bg = color.NRGBA{120, 120, 120, 80}
	}

	// Background rectangle
	bgRect := canvas.NewRectangle(bg)
	bgRect.CornerRadius = 8
	bgRect.SetMinSize(fyne.NewSize(0, 40))

	// The visible text
	label := canvas.NewText(msg, theme.Color(theme.ColorNameForeground))
	label.TextSize = 14
	label.Alignment = fyne.TextAlignLeading

	// Container variables must be declared before closure use
	var banner *fyne.Container

	// Close button
	closeBtn := widget.NewButton("×", func() {
		banner.Hide()
	})
	closeBtn.Importance = widget.LowImportance

	// Horizontal layout with some padding
	content := container.New(layout.NewHBoxLayout(),
		layout.NewSpacer(),
		label,
		layout.NewSpacer(),
		closeBtn,
	)
	content = container.NewPadded(content)

	// Stack: background first, then content
	banner = container.NewMax(bgRect, content)

	return banner
}

func main() {
	a := app.NewWithID("com.example.formsapp")
	w := a.NewWindow("Surveillance Forms")

	apiURL := "https://example.com/api/forms"
	appName := "forms-app"

	allForms, source, err := forms.LoadForms(a, apiURL, appName)
	if err != nil {
		fmt.Println("⚠️ Could not fetch from API:", err)
		allForms, err = forms.LoadFromEmbedded()
		source = "embedded"
		if err != nil {
			dialog.ShowError(fmt.Errorf("Failed to load any form definitions: %v", err), w)
			allForms = map[string]forms.FormDefinition{}
			source = "error"
		} else {
			fmt.Println("✅ Loaded embedded forms.json")
		}
	}

	// Create navigator
	nav := ui.NewNavigator(w)

	// --- Screen builders ---
	var loginScreen, verifyScreen, dashboardScreen func()

	loginScreen = func() {
		screen := ui.LoginScreen(a, func(phone string) {
			log.Println("Send verification code to:", phone)
			verifyScreen()
		})
		nav.Reset(screen)
	}

	verifyScreen = func() {
		content := ui.VerifyScreen(a, func(code string) {
			log.Println("Verified:", code)
			forms.StartAutoSync(a, "https://example.com/api/forms/submit")
			dashboardScreen()
			dashboardScreen()
		})
		back := widget.NewButton("← Back", func() { nav.PopSlide() })
		screen := container.NewBorder(back, nil, nil, nil, content)
		nav.PushSlide(screen)
	}

	dashboardScreen = func() {
		banner := statusBanner(source, w)
		content := ui.DashboardScreen(a, allForms, banner, func(name string) {
			formFields := allForms[name]
			formContent := forms.BuildForm(a, name, formFields.Sections, func(data map[string]string) {
				log.Println("Submitted", name, data)
				nav.PopSlide()
			})
			back := widget.NewButton("← Back", func() { nav.PopSlide() })
			screen := container.NewBorder(back, nil, nil, nil, formContent)
			nav.PushSlide(screen)
		})
		back := widget.NewButton("← Logout", func() { loginScreen() })
		screen := container.NewBorder(back, nil, nil, nil, content)
		nav.PushSlide(screen)
	}

	loginScreen()
	w.Resize(fyne.NewSize(400, 600))

	w.ShowAndRun()
}
