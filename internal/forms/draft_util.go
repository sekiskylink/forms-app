package forms

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
)

// LoadLatestDraft loads the most recent draft for the given form name.
func LoadLatestDraft(a fyne.App, formName string) (map[string]string, string) {
	root := a.Storage().RootURI().Path()
	if root == "" {
		home, _ := os.UserHomeDir()
		root = filepath.Join(home, ".forms-app")
	}
	draftDir := filepath.Join(root, "drafts")

	files, err := os.ReadDir(draftDir)
	if err != nil {
		return nil, ""
	}

	var latest os.DirEntry
	var latestTime int64
	for _, f := range files {
		if !strings.HasPrefix(f.Name(), formName+"-") || !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		parts := strings.Split(strings.TrimSuffix(f.Name(), ".json"), "-")
		if len(parts) < 2 {
			continue
		}
		if ts, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
			if ts > latestTime {
				latestTime = ts
				latest = f
			}
		}
	}
	if latest == nil {
		return nil, ""
	}

	path := filepath.Join(draftDir, latest.Name())
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, ""
	}

	var draft map[string]any
	if err := json.Unmarshal(data, &draft); err != nil {
		return nil, path
	}

	// unwrap nested structure
	payload := map[string]string{}
	if d, ok := draft["data"].(map[string]any); ok {
		for k, v := range d {
			payload[k] = fmt.Sprintf("%v", v)
		}
	}

	return payload, path
}
