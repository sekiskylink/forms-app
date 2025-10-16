package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// VerifyScreen asks for OTP code
func VerifyScreen(a fyne.App, onVerify func(code string)) fyne.CanvasObject {
	codeEntry := widget.NewEntry()
	codeEntry.SetPlaceHolder("Enter verification code")

	btn := widget.NewButton("Verify", func() {
		if codeEntry.Text != "" {
			onVerify(codeEntry.Text)
		}
	})

	return container.NewVBox(
		widget.NewLabelWithStyle("Verification", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		codeEntry,
		btn,
	)
}
