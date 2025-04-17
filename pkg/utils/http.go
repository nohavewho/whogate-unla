package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// MakeRequest makes an HTTP request with the given parameters
func MakeRequest(method, url string, headers map[string]string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	return client.Do(req)
}

// ReadResponseBody reads and unmarshals the response body
func ReadResponseBody(resp *http.Response, v any) error {
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

// WriteJSONResponse writes a JSON response
func WriteJSONResponse(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

// ParseJSONBody parses a JSON request body
func ParseJSONBody(r *http.Request, v any) error {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.Unmarshal(data, v)
}

// CopyHeaders copies headers from one request to another
func CopyHeaders(dst, src http.Header) {
	for k, v := range src {
		dst[k] = v
	}
}

// CloneRequest clones an HTTP request
func CloneRequest(r *http.Request) *http.Request {
	clone := r.Clone(r.Context())
	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err == nil {
			clone.Body = io.NopCloser(bytes.NewReader(body))
			r.Body = io.NopCloser(bytes.NewReader(body))
		}
	}
	return clone
}
