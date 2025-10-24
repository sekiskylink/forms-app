package ui

import (
	"bytes"
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"forms-app/internal/forms"
)

// DashboardScreen lists available forms and adds an offline drafts uploader.
func DashboardScreen(
	a fyne.App,
	formDefs map[string]forms.FormDefinition,
	banner fyne.CanvasObject,
	openForm func(name string),
) fyne.CanvasObject {

	// --- Build form cards ---
	var cards []fyne.CanvasObject
	for code, def := range formDefs {
		meta := def.Meta

		// Load icon
		var icon *canvas.Image
		if meta.Icon != "" {
			// Simple heuristic: URL vs local file
			if strings.HasPrefix(meta.Icon, "http://") || strings.HasPrefix(meta.Icon, "https://") {
				if uri, err := storage.ParseURI(meta.Icon); err == nil {
					icon = canvas.NewImageFromURI(uri)
				}
			} else {
				// Local/packaged file path
				// icon = canvas.NewImageFromFile(meta.Icon)
				switch meta.Icon {
				case "tb.png":
					icon = canvas.NewImageFromReader(bytes.NewReader(forms.IconTB), "tb")
				case "death.png":
					icon = canvas.NewImageFromReader(bytes.NewReader(forms.IconDeath), "death")
				case "cases.png":
					icon = canvas.NewImageFromReader(bytes.NewReader(forms.IconCases), "cases")
				default:
					res := fyne.CurrentApp().Settings().Theme().Icon(theme.IconNameFile)
					icon = canvas.NewImageFromResource(res)
				}
			}
		}

		if icon == nil {
			// ✅ Fyne v2.6.3-compatible themed fallback icon (single arg)
			res := a.Settings().Theme().Icon(theme.IconNameFile)
			icon = canvas.NewImageFromResource(res)
		}

		// consistent sizing
		icon.SetMinSize(fyne.NewSize(40, 40))
		icon.FillMode = canvas.ImageFillContain

		// Title + description (theme-friendly)
		title := widget.NewLabelWithStyle(meta.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		desc := widget.NewLabelWithStyle(meta.Description, fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
		desc.Wrapping = fyne.TextWrapWord

		textBox := container.NewVBox(title, desc)
		cardBody := container.NewBorder(nil, nil, icon, nil, textBox)

		// Clickable overlay
		btn := widget.NewButton("", func() { openForm(code) })
		btn.Importance = widget.LowImportance

		//card := container.NewStack(
		//	canvas.NewRectangle(color.NRGBA{245, 245, 245, 255}),
		//	container.NewPadded(cardBody),
		//	btn,
		//)
		padded := container.NewPadded(cardBody)
		card := NewHoverCard(padded, func() { openForm(code) })
		cards = append(cards, card)
	}

	if len(cards) == 0 {
		cards = append(cards, widget.NewLabel("⚠️ No forms available."))
	}

	// Scrollable list
	scroll := container.NewVScroll(container.NewVBox(cards...))
	scroll.SetMinSize(fyne.NewSize(360, 480))

	// --- Drafts ---
	apiURL := "https://example.com/api/forms/submit"
	drafts, _ := forms.LoadDrafts(a)

	var draftBtn fyne.CanvasObject
	if len(drafts) > 0 {
		count := len(drafts)
		draftBtn = widget.NewButton(fmt.Sprintf("📂 View %d Draft%s", count, plural(count)), func() {
			screen := DraftsScreen(a, apiURL, a.Driver().AllWindows()[0], func() {
				main := DashboardScreen(a, formDefs, banner, openForm)
				a.Driver().AllWindows()[0].SetContent(main)
			})
			a.Driver().AllWindows()[0].SetContent(screen)
		})
	} else {
		draftBtn = widget.NewLabel("No pending drafts")
	}

	// --- Theme toggle ---
	dark := a.Preferences().BoolWithFallback("darkMode", false)
	var themeBtn *widget.Button
	themeBtn = widget.NewButton("", func() {
		dark = !dark
		a.Preferences().SetBool("darkMode", dark)
		SetDark(a, dark)
		if dark {
			themeBtn.SetText("🌙 Dark")
		} else {
			themeBtn.SetText("🌞 Light")
		}
	})

	SetDark(a, dark)
	if dark {
		themeBtn.SetText("🌙 Dark")
	} else {
		themeBtn.SetText("🌞 Light")
	}

	// --- Auto sync + manual sync ---
	autoSwitch := widget.NewCheck("Auto Sync", func(on bool) {
		a.Preferences().SetBool("autoSyncEnabled", on)
		status := "disabled"
		if on {
			status = "enabled"
		}
		dialog.ShowInformation("Auto Sync", fmt.Sprintf("Auto sync %s.", status), a.Driver().AllWindows()[0])
	})
	autoSwitch.SetChecked(a.Preferences().BoolWithFallback("autoSyncEnabled", true))

	syncNowBtn := widget.NewButton("🔄 Sync Now", func() {
		go func() {
			fyne.Do(func() {
				dialog.ShowInformation("Manual Sync", "Starting draft sync...", a.Driver().AllWindows()[0])
			})
			success, failed := forms.ManualSync(a, apiURL)
			msg := fmt.Sprintf("✅ %d uploaded, ❌ %d failed", success, failed)
			fyne.Do(func() {
				dialog.ShowInformation("Sync Complete", msg, a.Driver().AllWindows()[0])
			})
		}()
	})
	syncNowBtn.Disable()
	if !autoSwitch.Checked {
		syncNowBtn.Enable()
	}
	autoSwitch.OnChanged = func(on bool) {
		a.Preferences().SetBool("autoSyncEnabled", on)
		if on {
			syncNowBtn.Disable()
		} else {
			syncNowBtn.Enable()
		}
	}

	// --- Header ---
	header := container.NewHBox(
		widget.NewLabelWithStyle("Available Forms", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		themeBtn,
		autoSwitch,
		syncNowBtn,
	)

	// --- Main content ---
	mainContent := container.NewVBox(
		header,
		draftBtn,
		container.NewPadded(scroll),
	)

	if banner != nil {
		return container.NewBorder(
			banner,
			nil, nil, nil,
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
