package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const BaseURL = "https://example.com/api" // placeholder

func PostJSON(endpoint string, payload any) (map[string]any, error) {
	b, _ := json.Marshal(payload)
	resp, err := http.Post(fmt.Sprintf("%s/%s", BaseURL, endpoint), "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result map[string]any
	err = json.Unmarshal(body, &result)
	return result, err
}
