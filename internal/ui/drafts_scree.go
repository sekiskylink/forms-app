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
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"forms-app/internal/forms"
)

// DraftsScreen displays saved drafts, allowing users to retry, preview, or delete them.
// The list automatically refreshes after retry or delete.
func DraftsScreen(a fyne.App, apiURL string, w fyne.Window, back func()) fyne.CanvasObject {
	listContainer := container.NewVBox()
	var refreshList func()

	refreshList = func() {
		listContainer.Objects = nil
		drafts, _ := forms.LoadDrafts(a)

		if len(drafts) == 0 {
			listContainer.Add(widget.NewLabelWithStyle(
				"No pending drafts.",
				fyne.TextAlignCenter,
				fyne.TextStyle{Italic: true},
			))
			listContainer.Refresh()
			return
		}

		// Retry All
		retryAllBtn := widget.NewButton(fmt.Sprintf("📤 Retry All (%d)", len(drafts)), func() {
			if len(drafts) == 0 {
				dialog.ShowInformation("No Drafts", "There are no drafts to upload.", w)
				return
			}

			progress := widget.NewProgressBar()
			progress.Min = 0
			progress.Max = float64(len(drafts))
			progress.SetValue(0)

			status := widget.NewLabel("Starting upload...")

			content := container.NewVBox(
				widget.NewLabel("Re-submitting all saved drafts…"),
				progress,
				status,
			)

			progressDialog := dialog.NewCustomWithoutButtons("Syncing Drafts", content, w)
			progressDialog.Show()

			var success, failed int

			go func() {
				for i, d := range drafts {
					status.SetText(fmt.Sprintf("Uploading %d of %d: %s", i+1, len(drafts), filepath.Base(d)))
					canvas.Refresh(status)

					err := forms.RetryDraft(a, apiURL, d)
					if err != nil {
						failed++
						fmt.Printf("❌ %s → %v\n", d, err)
					} else {
						success++
						fmt.Printf("✅ %s uploaded successfully\n", d)
					}

					progress.SetValue(float64(i + 1))
					canvas.Refresh(progress)
				}

				fyne.Do(func() {
					progressDialog.Hide()

					msg := fmt.Sprintf("✅ %d uploaded successfully\n❌ %d failed", success, failed)
					dialog.ShowInformation("Draft Sync Complete", msg, w)
					fyne.CurrentApp().SendNotification(&fyne.Notification{
						Title:   "Draft Sync Complete",
						Content: msg,
					})

					refreshList()
				})
			}()
		})

		listContainer.Add(retryAllBtn)
		listContainer.Add(widget.NewSeparator())

		for _, d := range drafts {
			base := filepath.Base(d)
			formName := strings.SplitN(base, "-", 2)[0]
			timestamp := extractTimestamp(base)
			content := fmt.Sprintf("%s (%s)", formName, timestamp)

			// Preview Button
			previewBtn := widget.NewButton("👁 Preview", func(path string) func() {
				return func() {
					data, err := os.ReadFile(path)
					if err != nil {
						dialog.ShowError(fmt.Errorf("Failed to read draft: %v", err), w)
						return
					}

					// Try pretty-printing JSON
					var pretty map[string]any
					jsonErr := json.Unmarshal(data, &pretty)
					var formatted []byte
					if jsonErr == nil {
						formatted, _ = json.MarshalIndent(pretty, "", "  ")
					} else {
						formatted = data // fallback to raw text
					}

					label := widget.NewLabel(string(formatted))
					label.TextStyle = fyne.TextStyle{Monospace: true}
					label.Wrapping = fyne.TextWrapWord

					scroll := container.NewVScroll(label)
					scroll.SetMinSize(fyne.NewSize(500, 400))

					// ✅ Only show Copy if JSON parsed successfully
					var content fyne.CanvasObject
					if jsonErr == nil {
						copyBtn := widget.NewButton("📋 Copy", func() {
							w.Clipboard().SetContent(string(formatted))
							fyne.CurrentApp().SendNotification(&fyne.Notification{
								Title:   "Copied",
								Content: "JSON copied to clipboard!",
							})
						})

						footer := container.NewHBox(layout.NewSpacer(), copyBtn, layout.NewSpacer())
						content = container.NewBorder(nil, footer, nil, nil, scroll)
					} else {
						content = scroll
					}

					// 🔹 Built-in “Close” button handles dismissal
					dialog.ShowCustom("Draft Preview", "Close", content, w)
				}
			}(d))

			retryBtn := widget.NewButton("🔄 Retry Upload", func(path string) func() {
				return func() {
					apiURL := "https://example.com/api/forms/submit"
					go func() {
						err := forms.RetryDraft(a, apiURL, path)
						fyne.Do(func() {
							if err != nil {
								dialog.ShowError(fmt.Errorf("Retry failed: %v", err), w)
							} else {
								dialog.ShowInformation("Success", "Draft uploaded successfully!", w)
								refreshList() // reload drafts after deletion
							}
						})
					}()
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
								refreshList()
							}
						}
					}, w)
					confirm.Show()
				}
			}(d))

			row := container.NewBorder(
				nil, nil,
				nil,
				container.NewHBox(previewBtn, retryBtn, deleteBtn),
				widget.NewLabel(content),
			)
			listContainer.Add(row)
		}
		listContainer.Refresh()
	}

	refreshList() // first render

	scroll := container.NewVScroll(listContainer)
	scroll.SetMinSize(fyne.NewSize(320, 480))

	return container.NewBorder(
		widget.NewButton("← Back", func() { back() }),
		nil, nil, nil,
		scroll,
	)
}

// extractTimestamp parses the timestamp from filename like CASES-1739950800.json → "2025-10-20 07:00".
func extractTimestamp(filename string) string {
	base := filepath.Base(filename)
	parts := strings.Split(base, "-")
	if len(parts) < 2 {
		return "unknown"
	}
	tsPart := strings.TrimSuffix(parts[1], ".json")
	if unix, err := strconv.ParseInt(tsPart, 10, 64); err == nil {
		t := time.Unix(unix, 0)
		return t.Format("2006-01-02 15:04")
	}
	return "unknown"
}
