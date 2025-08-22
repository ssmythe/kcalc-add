package handler

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddHandler_OK(t *testing.T) {
	t.Parallel()

	body := map[string]any{"a": 2.5, "b": 3.1}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/add", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	AddHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d, want %d", resp.StatusCode, http.StatusOK)
	}

	ct := resp.Header.Get("Content-Type")
	if ct == "" || ct[:16] != "application/json" {
		t.Fatalf("missing/incorrect Content-Type, got %q", ct)
	}

	var out struct {
		Result float64 `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}

	const eps = 1e-9
	want := 5.6
	if math.Abs(out.Result-want) > eps {
		t.Fatalf("result=%v, want %v", out.Result, want)
	}
}

func TestAddHandler_BadJSON(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/add", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	AddHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status=%d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestAddHandler_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/add", nil)
	w := httptest.NewRecorder()

	AddHandler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("status=%d, want %d", resp.StatusCode, http.StatusMethodNotAllowed)
	}
}
