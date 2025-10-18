package forms

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
)

func LoadForms(a fyne.App, apiURL, appName string) (map[string][]Section, string, error) {
	cache, _ := loadBundleFromCache(a, appName)

	// Try API first
	serverBundle, err := fetchBundleFromAPI(apiURL)
	if err != nil {
		if cache.Forms != nil {
			fmt.Println("⚠️ Using cached forms:", err)
			return cache.Forms, "cache", nil
		}
		return nil, "error", fmt.Errorf("no network and no cached forms available")
	}

	// Compare versions
	if cache.Version != "" && cache.Version == serverBundle.Version {
		fmt.Println("✅ Forms up to date (version", cache.Version, ")")
		return cache.Forms, "cache", nil
	}

	// Save new bundle
	if err := saveBundleToCache(a, appName, serverBundle); err != nil {
		fmt.Println("⚠️ Failed to update cache:", err)
	}
	fmt.Println("⬇️  Updated forms cache to version", serverBundle.Version)

	return serverBundle.Forms, "api", nil
}

// ------------------- helpers -------------------

func fetchBundleFromAPI(url string) (FormBundle, error) {
	var bundle FormBundle
	if url == "" {
		return bundle, errors.New("no API URL provided")
	}
	resp, err := http.Get(url)
	if err != nil {
		return bundle, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return bundle, fmt.Errorf("server returned %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&bundle); err != nil {
		return bundle, err
	}
	return bundle, nil
}

func saveBundleToCache(a fyne.App, appName string, bundle FormBundle) error {
	b, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return err
	}
	path, err := cachePath(a, appName)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(path, b, 0644)
}

func loadBundleFromCache(a fyne.App, appName string) (FormBundle, error) {
	path, err := cachePath(a, appName)
	if err != nil {
		return FormBundle{}, err
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return FormBundle{}, err
	}
	var bundle FormBundle
	err = json.Unmarshal(b, &bundle)
	return bundle, err
}

// cachePath returns platform-appropriate path for forms.json
func cachePath(a fyne.App, appName string) (string, error) {
	// On mobile, use Fyne storage
	if fyne.CurrentDevice().IsMobile() {
		root := a.Storage().RootURI()
		if root == nil {
			return "", errors.New("no app storage available")
		}
		path := filepath.Join(root.Path(), "forms.json")
		return path, nil
	}

	// On desktop, use config dir
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appName, "forms.json"), nil
}
