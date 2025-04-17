package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMakeRequest(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected method GET, got %s", r.Method)
		}
		if r.Header.Get("X-Test") != "test" {
			t.Errorf("expected header X-Test=test, got %s", r.Header.Get("X-Test"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Make request
	resp, err := MakeRequest("GET", ts.URL, map[string]string{"X-Test": "test"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestReadResponseBody(t *testing.T) {
	// Create test data
	data := map[string]string{"test": "value"}
	body, _ := json.Marshal(data)

	// Create response
	resp := &http.Response{
		Body: io.NopCloser(bytes.NewReader(body)),
	}

	// Read response
	var result map[string]string
	if err := ReadResponseBody(resp, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["test"] != "value" {
		t.Errorf("expected value 'value', got %s", result["test"])
	}
}

func TestWriteJSONResponse(t *testing.T) {
	// Create test data
	data := map[string]string{"test": "value"}

	// Create response writer
	w := httptest.NewRecorder()

	// Write response
	if err := WriteJSONResponse(w, http.StatusOK, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["test"] != "value" {
		t.Errorf("expected value 'value', got %s", result["test"])
	}
}

func TestParseJSONBody(t *testing.T) {
	// Create test data
	data := map[string]string{"test": "value"}
	body, _ := json.Marshal(data)

	// Create request
	r := httptest.NewRequest("POST", "/", bytes.NewReader(body))

	// Parse body
	var result map[string]string
	if err := ParseJSONBody(r, &result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["test"] != "value" {
		t.Errorf("expected value 'value', got %s", result["test"])
	}
}

func TestCopyHeaders(t *testing.T) {
	// Create source headers
	src := http.Header{}
	src.Set("X-Test", "test")

	// Create destination headers
	dst := http.Header{}

	// Copy headers
	CopyHeaders(dst, src)

	if dst.Get("X-Test") != "test" {
		t.Errorf("expected header X-Test=test, got %s", dst.Get("X-Test"))
	}
}

func TestCloneRequest(t *testing.T) {
	// Create test data
	data := map[string]string{"test": "value"}
	body, _ := json.Marshal(data)

	// Create request
	r := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	r.Header.Set("X-Test", "test")

	// Clone request
	clone := CloneRequest(r)

	// Check clone
	if clone.Method != r.Method {
		t.Errorf("expected method %s, got %s", r.Method, clone.Method)
	}
	if clone.URL.String() != r.URL.String() {
		t.Errorf("expected URL %s, got %s", r.URL.String(), clone.URL.String())
	}
	if clone.Header.Get("X-Test") != r.Header.Get("X-Test") {
		t.Errorf("expected header X-Test=%s, got %s", r.Header.Get("X-Test"), clone.Header.Get("X-Test"))
	}

	// Read original body
	var original map[string]string
	if err := ParseJSONBody(r, &original); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read clone body
	var cloned map[string]string
	if err := ParseJSONBody(clone, &cloned); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if original["test"] != cloned["test"] {
		t.Errorf("expected body value %s, got %s", original["test"], cloned["test"])
	}
}
