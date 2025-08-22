package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ssmythe/kcalc-add/handler"
)

func TestMain_AddEndpoint(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/add", handler.AddHandler)

	// spin up test server
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// make a request
	reqBody := []byte(`{"a":2.5,"b":3.1}`)
	resp, err := http.Post(ts.URL+"/add", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		t.Fatalf("http.Post error: %v", err)
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
		t.Errorf("got result=%v, want 5.6", out.Result)
	}
}
