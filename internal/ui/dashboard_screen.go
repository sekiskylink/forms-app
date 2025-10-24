package ui

import (
	"bytes"
	"fmt"
	"image/color"
	"strings"
	"time"

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

		var icon *canvas.Image
		if meta.Icon != "" {
			if strings.HasPrefix(meta.Icon, "http://") || strings.HasPrefix(meta.Icon, "https://") {
				if uri, err := storage.ParseURI(meta.Icon); err == nil {
					icon = canvas.NewImageFromURI(uri)
				}
			} else {
				switch meta.Icon {
				case "tb.png":
					icon = canvas.NewImageFromReader(bytes.NewReader(forms.IconTB), "tb")
				case "death.png":
					icon = canvas.NewImageFromReader(bytes.NewReader(forms.IconDeath), "death")
				case "cases.png":
					icon = canvas.NewImageFromReader(bytes.NewReader(forms.IconCases), "cases")
				default:
					res := a.Settings().Theme().Icon(theme.IconNameFile)
					icon = canvas.NewImageFromResource(res)
				}
			}
		}
		if icon == nil {
			res := a.Settings().Theme().Icon(theme.IconNameFile)
			icon = canvas.NewImageFromResource(res)
		}

		icon.SetMinSize(fyne.NewSize(40, 40))
		icon.FillMode = canvas.ImageFillContain

		title := widget.NewLabelWithStyle(meta.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		desc := widget.NewLabelWithStyle(meta.Description, fyne.TextAlignLeading, fyne.TextStyle{Italic: true})
		desc.Wrapping = fyne.TextWrapWord

		textBox := container.NewVBox(title, desc)
		cardBody := container.NewBorder(nil, nil, icon, nil, textBox)
		padded := container.NewPadded(cardBody)

		card := NewHoverCard(padded, func() { openForm(code) })
		cards = append(cards, card)
	}

	if len(cards) == 0 {
		cards = append(cards, widget.NewLabel("⚠️ No forms available."))
	}

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
	var icon fyne.Resource
	if dark {
		icon = theme.NewThemedResource(theme.VisibilityIcon())
	} else {
		icon = theme.NewThemedResource(theme.VisibilityOffIcon())
	}
	themeBtn = widget.NewButtonWithIcon("", icon, func() {
		dark = !dark
		a.Preferences().SetBool("darkMode", dark)
		SetDark(a, dark)
		if dark {
			themeBtn.SetIcon(theme.NewThemedResource(theme.VisibilityIcon()))
		} else {
			themeBtn.SetIcon(theme.NewThemedResource(theme.VisibilityOffIcon()))
		}
	})

	// --- Sync buttons ---
	autoSwitch := widget.NewCheck("Auto Sync", func(on bool) {
		a.Preferences().SetBool("autoSyncEnabled", on)
		status := "disabled"
		if on {
			status = "enabled"
		}
		dialog.ShowInformation("Auto Sync", fmt.Sprintf("Auto sync %s.", status), a.Driver().AllWindows()[0])
	})
	autoSwitch.SetChecked(a.Preferences().BoolWithFallback("autoSyncEnabled", true))

	syncNowBtn := widget.NewButtonWithIcon("", theme.ViewRefreshIcon(), func() {
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

	// --- Side Drawer (non-navigating overlay) ---
	isDrawerOpen := false
	var sideDrawer fyne.CanvasObject
	var overlay fyne.CanvasObject
	win := a.Driver().AllWindows()[0]
	canvasWidth := win.Canvas().Size().Width

	drawerWidth := fyne.Min(300, canvasWidth*0.5)
	sideDrawer = buildSideDrawer(a, func() {
		main := LoginScreen(a, func(phone string) {
			fmt.Println("Logged out:", phone)
		})
		a.Driver().AllWindows()[0].SetContent(main)
	}, func() {
		// ✅ Close drawer callback
		if isDrawerOpen {
			hideDrawer(sideDrawer, overlay, drawerWidth)
			isDrawerOpen = false
		}
	})

	sideDrawer.Resize(fyne.NewSize(drawerWidth, win.Canvas().Size().Height))
	sideDrawer.Move(fyne.NewPos(-drawerWidth, 0))
	sideDrawer.Hide()

	overlay = newTappableOverlay(func() {
		if isDrawerOpen {
			hideDrawer(sideDrawer, overlay, drawerWidth)
			isDrawerOpen = false
		}
	})
	overlay.Hide()

	// ESC closes drawer
	win.Canvas().SetOnTypedKey(func(ev *fyne.KeyEvent) {
		if ev.Name == fyne.KeyEscape && isDrawerOpen {
			hideDrawer(sideDrawer, overlay, drawerWidth)
			isDrawerOpen = false
		}
	})

	// --- App Bar ---
	appBarColor := color.NRGBA{25, 118, 210, 255}
	appBarBg := canvas.NewRectangle(appBarColor)
	appBarBg.SetMinSize(fyne.NewSize(0, 56))
	titleLabel := widget.NewLabelWithStyle("Surveillance Forms", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	titleLabel.Alignment = fyne.TextAlignLeading

	menuBtn := widget.NewButton("☰", func() {
		if isDrawerOpen {
			hideDrawer(sideDrawer, overlay, drawerWidth)
		} else {
			showDrawer(sideDrawer, overlay, drawerWidth)
		}
		isDrawerOpen = !isDrawerOpen
	})

	appBarContent := container.NewHBox(menuBtn, layout.NewSpacer(), titleLabel, layout.NewSpacer(), themeBtn, syncNowBtn)
	appBar := container.NewMax(appBarBg, container.NewPadded(appBarContent))

	content := container.NewVBox(appBar, banner, draftBtn, container.NewPadded(scroll))

	// ✅ Drawer is purely an overlay, not part of navigation
	root := container.NewStack(content, overlay, sideDrawer)
	return root
}

func showDrawer(sideDrawer fyne.CanvasObject, overlay fyne.CanvasObject, width float32) {
	sideDrawer.Show()
	overlay.Show()
	anim := canvas.NewPositionAnimation(
		fyne.NewPos(-width, 0),
		fyne.NewPos(0, 0),
		200*time.Millisecond,
		func(p fyne.Position) {
			sideDrawer.Move(p)
			canvas.Refresh(sideDrawer)
		},
	)
	anim.Start()
}

func hideDrawer(sideDrawer fyne.CanvasObject, overlay fyne.CanvasObject, width float32) {
	anim := canvas.NewPositionAnimation(
		sideDrawer.Position(),
		fyne.NewPos(-width, 0),
		200*time.Millisecond,
		func(p fyne.Position) {
			sideDrawer.Move(p)
			canvas.Refresh(sideDrawer)
		},
	)
	anim.Start()

	// ⏳ Wait slightly longer than the animation duration, then hide both
	go func() {
		time.Sleep(220 * time.Millisecond) // a little buffer
		fyne.Do(func() {
			sideDrawer.Hide()
			overlay.Hide()
		})
	}()
}

func buildSideDrawer(a fyne.App, onLogout func(), closeDrawer func()) fyne.CanvasObject {
	bg := canvas.NewRectangle(color.NRGBA{255, 255, 255, 255})

	// Header
	title := widget.NewLabelWithStyle("Menu", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	closeBtn := widget.NewButton("✕", func() {
		closeDrawer() // ✅ invoke parent's close function
	})

	header := container.NewBorder(nil, nil, nil, closeBtn, title)

	// Menu items
	logoutBtn := widget.NewButtonWithIcon("Logout", theme.LogoutIcon(), func() {
		closeDrawer()
		onLogout()
	})
	settingsBtn := widget.NewButtonWithIcon("Settings", theme.SettingsIcon(), func() {
		dialog.ShowInformation("Settings", "Settings coming soon.", a.Driver().AllWindows()[0])
	})
	aboutBtn := widget.NewButtonWithIcon("About", theme.InfoIcon(), func() {
		dialog.ShowInformation("About", "Surveillance Forms v1.0", a.Driver().AllWindows()[0])
	})

	sideContent := container.NewVBox(
		header,
		widget.NewSeparator(),
		logoutBtn,
		settingsBtn,
		aboutBtn,
		layout.NewSpacer(),
	)
	side := container.NewMax(bg, container.NewPadded(sideContent))

	// ✅ Half-screen width
	win := a.Driver().AllWindows()[0]
	winWidth := win.Canvas().Size().Width
	drawerWidth := fyne.Min(300, winWidth*0.5)
	side.Resize(fyne.NewSize(drawerWidth, win.Canvas().Size().Height))
	side.Move(fyne.NewPos(-drawerWidth, 0)) // start hidden off-screen

	return side
}

// plural adds "s" for plural count
func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
