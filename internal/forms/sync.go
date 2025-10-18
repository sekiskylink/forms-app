package forms

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"fyne.io/fyne/v2"
)

var (
	autoSyncRunning bool
	autoSyncMutex   sync.Mutex
)

// StartAutoSync runs a background goroutine that periodically syncs drafts.
// It checks the preference "autoSyncEnabled" and only runs if true.
func StartAutoSync(a fyne.App, apiURL string) {
	autoSyncMutex.Lock()
	if autoSyncRunning {
		autoSyncMutex.Unlock()
		return // already running
	}
	autoSyncRunning = true
	autoSyncMutex.Unlock()

	go func() {
		for {
			time.Sleep(60 * time.Second)

			if !a.Preferences().BoolWithFallback("autoSyncEnabled", true) {
				continue // user disabled it
			}

			if !IsOnline() {
				continue
			}

			root := a.Storage().RootURI().Path()
			if root == "" {
				home, err := os.UserHomeDir()
				if err != nil {
					continue
				}
				root = filepath.Join(home, ".forms-app")
			}
			draftDir := filepath.Join(root, "drafts")
			entries, err := os.ReadDir(draftDir)
			if err != nil || len(entries) == 0 {
				continue
			}

			success, failed := 0, 0
			for _, e := range entries {
				if e.IsDir() {
					continue
				}
				path := filepath.Join(draftDir, e.Name())
				if err := RetryDraft(a, apiURL, path); err != nil {
					failed++
				} else {
					success++
				}
			}

			if success > 0 || failed > 0 {
				summary := fmt.Sprintf("Auto-sync complete: ✅ %d uploaded, ❌ %d failed", success, failed)
				fmt.Println(summary)
				fyne.CurrentApp().SendNotification(&fyne.Notification{
					Title:   "Auto Sync Complete",
					Content: summary,
				})
			}
		}
	}()
}

// ManualSync tries to upload all drafts immediately.
// Returns (successCount, failedCount)
func ManualSync(a fyne.App, apiURL string) (int, int) {
	root := a.Storage().RootURI().Path()
	if root == "" {
		home, _ := os.UserHomeDir()
		root = filepath.Join(home, ".forms-app")
	}
	draftDir := filepath.Join(root, "drafts")

	entries, err := os.ReadDir(draftDir)
	if err != nil {
		return 0, 0
	}
	success, failed := 0, 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		path := filepath.Join(draftDir, e.Name())
		if err := RetryDraft(a, apiURL, path); err != nil {
			failed++
		} else {
			success++
		}
	}
	return success, failed
}

// IsOnline checks whether we have an active internet connection.
func IsOnline() bool {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://clients3.google.com/generate_204")
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode == 204
}

// getDraftDir resolves the current app's draft directory (cross-platform)
func getDraftDir(a fyne.App) string {
	root := a.Storage().RootURI().Path()
	if root == "" {
		home, _ := os.UserHomeDir()
		root = filepath.Join(home, ".forms-app")
	}
	draftDir := filepath.Join(root, "drafts")
	os.MkdirAll(draftDir, 0755)
	return draftDir
}
