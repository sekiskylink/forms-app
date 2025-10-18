package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"forms-app/internal/forms"
)

// DraftsScreen shows all locally saved drafts with options to retry or delete them.
func DraftsScreen(a fyne.App, apiURL string, w fyne.Window, back func()) fyne.CanvasObject {
	drafts, _ := forms.LoadDrafts(a)

	if len(drafts) == 0 {
		return container.NewVBox(
			widget.NewLabelWithStyle("No pending drafts.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
			widget.NewButton("← Back", func() { back() }),
		)
	}

	var items []fyne.CanvasObject

	// Retry All button
	retryAllBtn := widget.NewButton(fmt.Sprintf("📤 Retry All (%d)", len(drafts)), func() {
		var success, failed int
		for _, d := range drafts {
			if err := forms.RetryDraft(a, apiURL, d); err != nil {
				failed++
			} else {
				success++
			}
		}
		msg := fmt.Sprintf("✅ %d re-submitted, ❌ %d failed.", success, failed)
		dialog.ShowInformation("Draft Sync", msg, w)
		back()
	})
	items = append(items, retryAllBtn)

	items = append(items, widget.NewSeparator())

	// Each draft entry
	for _, d := range drafts {
		base := filepath.Base(d)

		formName := strings.SplitN(base, "-", 2)[0]
		timestamp := extractTimestamp(base)
		data, _ := os.ReadFile(d)
		var draft map[string]any
		_ = json.Unmarshal(data, &draft)

		var errLabel string
		if e, ok := draft["error"].(map[string]any); ok {
			switch e["type"] {
			case "network":
				errLabel = "Network Error"
			case "server":
				code, _ := e["code"].(float64)
				msg, _ := e["message"].(string)
				if code > 0 {
					errLabel = fmt.Sprintf("Server Error %d", int(code))
				} else if msg != "" {
					errLabel = fmt.Sprintf("Server Error: %s", msg)
				}
			default:
				errLabel = "Unknown Error"
			}
		}

		content := fmt.Sprintf("%s  (%s) – %s", formName, timestamp, errLabel)

		retryBtn := widget.NewButton("🔁 Retry", func(path string) func() {
			return func() {
				if err := forms.RetryDraft(a, apiURL, path); err != nil {
					dialog.ShowError(fmt.Errorf("Retry failed: %v", err), w)
				} else {
					dialog.ShowInformation("✅ Success", "Draft re-submitted successfully!", w)
					back()
				}
			}
		}(d))

		deleteBtn := widget.NewButton("🗑 Delete", func(path string) func() {
			return func() {
				confirm := dialog.NewConfirm("Delete Draft", "Are you sure you want to delete this draft?", func(yes bool) {
					if yes {
						if err := forms.DeleteDraft(a, path); err != nil {
							dialog.ShowError(err, w)
						} else {
							dialog.ShowInformation("Deleted", "Draft removed.", w)
							back()
						}
					}
				}, w)
				confirm.Show()
			}
		}(d))

		row := container.NewBorder(nil, nil,
			nil, container.NewHBox(retryBtn, deleteBtn),
			widget.NewLabel(content),
		)
		items = append(items, row)
	}

	scroll := container.NewVScroll(container.NewVBox(items...))
	return container.NewBorder(
		widget.NewButton("← Back", func() { back() }),
		nil, nil, nil,
		scroll,
	)
}

// extractTimestamp parses the timestamp portion from filename
func extractTimestamp(filename string) string {
	base := filepath.Base(filename)
	parts := strings.Split(base, "-")
	if len(parts) < 2 {
		return "unknown"
	}

	tsPart := strings.TrimSuffix(parts[1], ".json")

	// Try parse as int64 (Unix timestamp)
	if unix, err := strconv.ParseInt(tsPart, 10, 64); err == nil {
		t := time.Unix(unix, 0)
		return t.Format("2006-01-02 15:04")
	}

	return "unknown"
}
