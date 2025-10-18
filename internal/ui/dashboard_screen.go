package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
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
		n := name
		btn := widget.NewButton("📝 "+n, func() { openForm(n) })
		buttons = append(buttons, btn)
	}

	if len(buttons) == 0 {
		buttons = append(buttons, widget.NewLabel("⚠️ No forms available."))
	}

	scroll := container.NewVScroll(container.NewVBox(buttons...))
	scroll.SetMinSize(fyne.NewSize(300, 400))

	apiURL := "https://example.com/api/forms/submit" // your real endpoint
	drafts, _ := forms.LoadDrafts(a)

	// 🗂 Draft button (only if drafts exist)
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

	// 🌐 Auto-Sync toggle
	autoSwitch := widget.NewCheck("Auto Sync", func(on bool) {
		a.Preferences().SetBool("autoSyncEnabled", on)
		status := "disabled"
		if on {
			status = "enabled"
		}
		dialog.ShowInformation("Auto Sync", fmt.Sprintf("Auto sync %s.", status), a.Driver().AllWindows()[0])
	})
	autoSwitch.SetChecked(a.Preferences().BoolWithFallback("autoSyncEnabled", true))

	// ⚙️ Manual “Sync Now” button (only visible when auto-sync is off)
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

	header := container.NewHBox(
		widget.NewLabelWithStyle("Available Forms", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		layout.NewSpacer(),
		autoSwitch,
		syncNowBtn,
	)

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
} // plural adds "s" for plural count
func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
