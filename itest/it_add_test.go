// itest/it_add_test.go
package itest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"testing"
)

func baseURL(t *testing.T) string {
	t.Helper()
	if v := os.Getenv("BASE_URL"); v != "" {
		return v
	}
	t.Skip("BASE_URL not set; skipping integration test")
	return ""
}

func TestAdd_Live(t *testing.T) {
	url := baseURL(t) + "/add"

	req := []byte(`{"a":2.5,"b":3.1}`)
	resp, err := http.Post(url, "application/json", bytes.NewReader(req))
	if err != nil {
		t.Fatalf("POST /add error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d, want 200", resp.StatusCode)
	}

	var out struct {
		Result float64 `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.Result != 5.6 {
		t.Fatalf("result=%v, want 5.6", out.Result)
	}
}
