package forms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
)

func SubmitForm(a fyne.App, apiURL, formName string, payload map[string]string) error {
	client := &http.Client{Timeout: 15 * time.Second}

	body, err := json.Marshal(map[string]any{
		"form": formName,
		"data": payload,
	})
	if err != nil {
		return fmt.Errorf("marshal error: %v", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("request error: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		// 🟡 Network-level failure
		draft := map[string]any{
			"form":      formName,
			"data":      payload,
			"error":     map[string]any{"type": "network", "message": err.Error()},
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
		_ = saveTaggedDraft(a, formName, draft)
		return fmt.Errorf("offline mode — form saved for later upload")
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	// ✅ Success
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	// 🔴 Any non-successful response → save tagged draft
	draft := map[string]any{
		"form":      formName,
		"data":      payload,
		"error":     map[string]any{"type": "server", "code": resp.StatusCode, "message": string(respBody)},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	_ = saveTaggedDraft(a, formName, draft)

	return fmt.Errorf("submission failed (%d): %s — form saved locally", resp.StatusCode, string(respBody))
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

// RetryDraft re-submits a stored draft and deletes it on success.
func RetryDraft(a fyne.App, apiURL, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var payload map[string]string
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	formName := filepath.Base(path)
	formName = strings.TrimSuffix(formName, filepath.Ext(formName))

	if err := SubmitForm(a, apiURL, formName, payload); err != nil {
		return err
	}

	return os.Remove(path)
}

// saveTaggedDraft stores a rich JSON draft with metadata like error type and timestamp.
func saveTaggedDraft(a fyne.App, formName string, payload map[string]any) error {
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
