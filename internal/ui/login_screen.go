package ui

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// LoginScreen shows phone entry for authentication
func LoginScreen2(a fyne.App, onNext func(phone string)) fyne.CanvasObject {
	phoneEntry := widget.NewEntry()
	phoneEntry.SetPlaceHolder("Enter phone number")
	updating := false
	phoneEntry.OnChanged = func(text string) {
		if updating {
			return
		}
		updating = true

		// Keep only digits and at most one '+' at the beginning
		var b strings.Builder
		for i, r := range text {
			if r >= '0' && r <= '9' {
				b.WriteRune(r)
			} else if r == '+' && i == 0 {
				b.WriteRune(r)
			}
			// ignore everything else (letters, punctuation, spaces, etc.)
		}

		clean := b.String()
		if clean != text {
			phoneEntry.SetText(clean)
		}

		updating = false
	}

	btn := widget.NewButton("Send Code", func() {
		if phoneEntry.Text != "" {
			onNext(phoneEntry.Text)
		}
	})

	return container.NewVBox(
		widget.NewLabelWithStyle("User Login", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		phoneEntry,
		btn,
	)
}

func LoginScreen(a fyne.App, onSubmit func(phone string)) fyne.CanvasObject {
	// === Gradient background ===
	top := color.NRGBA{0, 32, 96, 255}
	bottom := color.NRGBA{48, 128, 255, 255}
	bg := canvas.NewLinearGradient(top, bottom, 90)

	// === Logo area (about 45%) ===
	circle := canvas.NewCircle(color.NRGBA{255, 255, 255, 40})
	circle.Resize(fyne.NewSize(120, 120))

	icon := widget.NewIcon(theme.MailSendIcon()) // Replace with your logo resource
	icon.Resize(fyne.NewSize(60, 60))
	logo := container.NewCenter(container.NewMax(circle, icon))

	// App name — now white
	appName := canvas.NewText("SukumaPro", color.White)
	appName.Alignment = fyne.TextAlignCenter
	appName.TextStyle = fyne.TextStyle{Bold: true}
	appName.TextSize = 24

	logoArea := container.NewVBox(
		layout.NewSpacer(),
		container.NewCenter(logo),
		container.NewCenter(appName),
		layout.NewSpacer(),
	)
	logoArea = container.NewPadded(logoArea)

	// === Login form ===
	loginTitle := widget.NewLabelWithStyle("Login", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	entry := widget.NewEntry()
	entry.SetPlaceHolder("Mobile Number")

	phoneIcon := canvas.NewText("📞", theme.Color(theme.ColorNameForeground))
	phoneIcon.TextSize = 20
	phoneIcon.Alignment = fyne.TextAlignCenter
	// entryIcon := widget.NewIcon(theme.PhoneIcon())
	entryRow := container.NewBorder(nil, nil, phoneIcon, nil, entry)
	updating := false

	entry.OnChanged = func(text string) {
		if updating {
			return
		}
		updating = true

		var b strings.Builder
		for i, r := range text {
			switch {
			case r >= '0' && r <= '9':
				b.WriteRune(r)
			case r == '+' && i == 0:
				// allow '+' only if it’s the first rune and not already present
				if !strings.ContainsRune(text[:i], '+') {
					b.WriteRune(r)
				}
				// ignore everything else (letters, symbols, spaces, etc.)
			}
		}

		clean := b.String()
		if clean != text {
			entry.SetText(clean)
		}

		updating = false
	}

	loginBtn := NewAccentButton("Login", func() {
		if onSubmit != nil {
			onSubmit(entry.Text)
		}
	})
	// loginBtn.Importance = widget.HighImportance
	loginBtn.Resize(fyne.NewSize(200, 44))
	copyright := fmt.Sprintf("© 2019–%d Sekiwere Samuel", time.Now().Year())

	footer := widget.NewLabelWithStyle(
		copyright,
		fyne.TextAlignCenter,
		fyne.TextStyle{Italic: true},
	)

	form := container.NewVBox(
		loginTitle,
		layout.NewSpacer(),
		entryRow,
		widget.NewSeparator(),
		loginBtn,
		layout.NewSpacer(),
		footer,
	)
	form = container.NewPadded(form)

	// === White bottom sheet (floating) ===
	cardBG := canvas.NewRectangle(color.NRGBA{255, 255, 255, 255})
	cardBG.CornerRadius = 32 // only top corners rounded
	cardBG.SetMinSize(fyne.NewSize(0, 400))

	// Soft drop shadow (light gray, slightly offset)
	shadow := canvas.NewRectangle(color.NRGBA{0, 0, 0, 25})
	shadow.SetMinSize(fyne.NewSize(0, 410))
	shadow.Move(fyne.NewPos(0, 5))

	card := container.NewStack(shadow, cardBG, form)

	// Slight upward offset to overlap gradient (≈ 20 px)
	overlapContainer := container.NewVBox(
		layout.NewSpacer(),
		container.NewPadded(card),
	)
	overlapContainer.Move(fyne.NewPos(0, -20))

	// === Combine ===
	content := container.NewStack(
		bg,
		container.NewVBox(
			container.NewPadded(logoArea),
			layout.NewSpacer(),
			overlapContainer,
		),
	)

	return content
}
