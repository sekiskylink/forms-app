package forms

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
)

// SubmitForm sends a filled form to the backend API.
// If the network is unreachable or the server returns an error,
// it saves the payload locally as a standardized draft.
func SubmitForm(a fyne.App, apiURL, formName string, payload map[string]string) error {
	// Prepare JSON body for submission
	body, err := json.Marshal(map[string]any{
		"form": formName,
		"data": payload,
	})
	if err != nil {
		return fmt.Errorf("marshal error: %v", err)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("request creation error: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform the network request
	resp, err := client.Do(req)
	if err != nil {
		// 🟡 Network failure → store as standardized draft
		draft := map[string]any{
			"form": formName,
			"data": payload,
			"meta": map[string]any{
				"saved_at": time.Now().UTC().Format(time.RFC3339),
				"source":   "auto",
			},
			"error": map[string]any{
				"type":    "network",
				"message": err.Error(),
			},
		}
		if saveErr := SaveTaggedDraft(a, formName, draft); saveErr != nil {
			return fmt.Errorf("network error: %v (and failed to save draft: %v)", err, saveErr)
		}
		return fmt.Errorf("offline mode — form saved to drafts for later upload")
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	// ✅ Success (200–299)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	// 🔴 Server-side error → save as standardized draft
	draft := map[string]any{
		"form": formName,
		"data": payload,
		"meta": map[string]any{
			"saved_at": time.Now().UTC().Format(time.RFC3339),
			"source":   "auto",
		},
		"error": map[string]any{
			"type":    "server",
			"message": fmt.Sprintf("status %d: %s", resp.StatusCode, string(respBody)),
		},
	}

	_ = SaveTaggedDraft(a, formName, draft)
	return errors.New(fmt.Sprintf("server error (%d): %s", resp.StatusCode, string(respBody)))
}

func saveDraft(a fyne.App, formName string, payload map[string]string) error {
	root := a.Storage().RootURI().Path()
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot resolve home dir: %v", err)
		}
		root = filepath.Join(home, ".forms-app")
	}

	draftDir := filepath.Join(root, "drafts")

	// Always ensure directory exists
	if err := os.MkdirAll(draftDir, 0755); err != nil {
		return fmt.Errorf("cannot create draft directory: %v", err)
	}

	filename := fmt.Sprintf("%s-%d.json", formName, time.Now().Unix())
	path := filepath.Join(draftDir, filename)

	data, _ := json.MarshalIndent(payload, "", "  ")

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to save draft: %v", err)
	}

	fmt.Printf("💾 Draft saved: %s\n", path) // <== helpful log

	return nil
}

// LoadDrafts lists saved drafts for later re-submission.
func LoadDrafts(a fyne.App) ([]string, error) {
	root := a.Storage().RootURI().Path()
	if root == "" {
		home, _ := os.UserHomeDir()
		root = filepath.Join(home, ".forms-app")
	}
	draftDir := filepath.Join(root, "drafts")

	files, err := os.ReadDir(draftDir)
	if err != nil {
		return nil, err
	}

	var drafts []string
	for _, f := range files {
		if !f.IsDir() {
			drafts = append(drafts, filepath.Join(draftDir, f.Name()))
		}
	}
	return drafts, nil
}

type Draft struct {
	Form  string            `json:"form"`
	Data  map[string]string `json:"data"`
	Meta  map[string]any    `json:"meta"`
	Error map[string]any    `json:"error"`
}

// RetryDraft tries to re-upload a saved draft file and deletes it on success.
func RetryDraft(a fyne.App, apiURL, draftPath string) error {
	data, err := os.ReadFile(draftPath)
	if err != nil {
		return fmt.Errorf("failed to read draft: %v", err)
	}

	var d Draft
	if err := json.Unmarshal(data, &d); err != nil {
		return fmt.Errorf("invalid draft format: %v", err)
	}

	if d.Form == "" || len(d.Data) == 0 {
		return fmt.Errorf("draft missing form or data fields")
	}

	// Try submission
	err = SubmitForm(a, apiURL, d.Form, d.Data)
	if err != nil {
		return fmt.Errorf("retry failed: %v", err)
	}

	// On success → delete file
	if rmErr := os.Remove(draftPath); rmErr != nil {
		return fmt.Errorf("submitted but failed to delete draft: %v", rmErr)
	}

	fyne.CurrentApp().SendNotification(&fyne.Notification{
		Title:   "✅ Draft Uploaded",
		Content: fmt.Sprintf("%s uploaded successfully and removed.", filepath.Base(draftPath)),
	})

	return nil
}

// SaveTaggedDraft stores a rich JSON draft with metadata like error type and timestamp.
func SaveTaggedDraft(a fyne.App, formName string, payload map[string]any) error {
	root := a.Storage().RootURI().Path()
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot resolve home dir: %v", err)
		}
		root = filepath.Join(home, ".forms-app")
	}

	draftDir := filepath.Join(root, "drafts")
	if err := os.MkdirAll(draftDir, 0755); err != nil {
		return fmt.Errorf("cannot create draft directory: %v", err)
	}

	filename := fmt.Sprintf("%s-%d.json", formName, time.Now().Unix())
	path := filepath.Join(draftDir, filename)

	data, _ := json.MarshalIndent(payload, "", "  ")
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to save draft: %v", err)
	}

	fmt.Printf("💾 Draft saved: %s\n", path)
	return nil
}

// DeleteDraft removes a saved draft file.
func DeleteDraft(a fyne.App, path string) error {
	return os.Remove(path)
}
