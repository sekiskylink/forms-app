package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// LoginScreen shows phone entry for authentication
func LoginScreen(a fyne.App, onNext func(phone string)) fyne.CanvasObject {
	phoneEntry := widget.NewEntry()
	phoneEntry.SetPlaceHolder("Enter phone number")

	btn := widget.NewButton("Send Code", func() {
		if phoneEntry.Text != "" {
			onNext(phoneEntry.Text)
		}
	})

	return container.NewVBox(
		widget.NewLabelWithStyle("Login", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		phoneEntry,
		btn,
	)
}
