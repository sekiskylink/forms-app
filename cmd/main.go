package main

import (
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"forms-app/internal/forms"
	"forms-app/internal/ui"
)

func main() {
	a := app.New()
	w := a.NewWindow("Surveillance Forms")

	// Load form definitions
	wd, _ := os.Getwd()
	formPath := filepath.Join(wd, "../assets", "forms.json")
	allForms, _ := forms.LoadForms(formPath)

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
			dashboardScreen()
		})
		back := widget.NewButton("← Back", func() { nav.PopSlide() })
		screen := container.NewBorder(back, nil, nil, nil, content)
		nav.PushSlide(screen)
	}

	dashboardScreen = func() {
		content := ui.DashboardScreen(a, allForms, func(name string) {
			formFields := allForms[name]
			formContent := forms.BuildForm(a, name, formFields, func(data map[string]string) {
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

	// Start app
	loginScreen()
	w.Resize(fyne.NewSize(400, 600))
	w.ShowAndRun()
}
