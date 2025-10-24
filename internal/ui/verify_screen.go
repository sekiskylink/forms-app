package ui

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// VerifyScreen asks for OTP code
//func VerifyScreen(a fyne.App, onVerify func(code string)) fyne.CanvasObject {
//	codeEntry := widget.NewEntry()
//	codeEntry.SetPlaceHolder("Enter verification code")
//
//	btn := widget.NewButton("Verify", func() {
//		if codeEntry.Text != "" {
//			onVerify(codeEntry.Text)
//		}
//	})
//
//	return container.NewVBox(
//		widget.NewLabelWithStyle("Verification", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
//		codeEntry,
//		btn,
//	)
//}

func VerifyScreen(a fyne.App, onVerify func(code string)) fyne.CanvasObject {
	// --- Top banner / logo area ---
	bgColor := color.NRGBA{20, 90, 160, 255} // deep blue
	logoText := canvas.NewText("🔐 Verify Code", color.White)
	logoText.Alignment = fyne.TextAlignCenter
	logoText.TextSize = 24

	topArea := container.NewCenter(logoText)
	topAreaBG := canvas.NewRectangle(bgColor)
	topAreaBG.SetMinSize(fyne.NewSize(0, 180))
	top := container.NewStack(topAreaBG, topArea)

	// --- Entry field (digits only) ---
	codeEntry := widget.NewEntry()
	codeEntry.SetPlaceHolder("Enter verification code")
	codeEntry.Wrapping = fyne.TextWrapOff
	codeEntry.TextStyle = fyne.TextStyle{Monospace: true}

	codeEntry.OnChanged = func(text string) {
		clean := strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' {
				return r
			}
			return -1
		}, text)
		if clean != text {
			codeEntry.SetText(clean)
		}
		if len(clean) > 6 {
			codeEntry.SetText(clean[:6])
		}
	}

	// Wrap entry in a responsive box (expands with window)
	entryBox := container.New(layout.NewMaxLayout(),
		container.NewPadded(codeEntry),
	)

	// --- Buttons ---
	verifyBtn := widget.NewButtonWithIcon("Verify", theme.ConfirmIcon(), func() {
		code := strings.TrimSpace(codeEntry.Text)
		if len(code) < 4 {
			dialog.ShowInformation("Invalid Code", "Enter a valid 4–6 digit code.", a.Driver().AllWindows()[0])
			return
		}
		onVerify(code)
	})
	verifyBtn.Importance = widget.HighImportance

	resendBtn := widget.NewButtonWithIcon("Resend Code", theme.MailSendIcon(), func() {
		dialog.ShowInformation("Code Sent", "A new verification code has been sent.", a.Driver().AllWindows()[0])
	})

	// Buttons row expands evenly
	buttonsBox := container.NewGridWithColumns(2, verifyBtn, resendBtn)

	// --- Copyright ---
	copy := widget.NewLabelWithStyle("© 2025 SukumaPro", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	// --- Form card ---
	formBox := container.NewVBox(
		widget.NewLabelWithStyle("Verification", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		entryBox,
		buttonsBox,
		layout.NewSpacer(),
		copy,
	)
	formBox = container.NewPadded(formBox)

	cardBG := canvas.NewRectangle(color.White)
	cardBG.CornerRadius = 16
	card := container.NewStack(cardBG, formBox)

	// Responsive card width: center + margin padding
	cardContainer := container.NewHBox(
		layout.NewSpacer(),
		container.NewMax(card),
		layout.NewSpacer(),
	)

	// --- Combine top + card ---
	content := container.NewVBox(
		top,
		layout.NewSpacer(),
		cardContainer,
		layout.NewSpacer(),
	)

	return container.NewBorder(nil, nil, nil, nil, content)
}
