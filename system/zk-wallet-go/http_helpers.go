package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
)

// sha256hex computes SHA-256 of b and returns a URL-safe Base64 (no padding) string.
// NOTE: Despite the name "hex", this returns Base64URL, not hex. Consider renaming to sha256b64url.
func sha256hex(b []byte) string {
	h := sha256.Sum256(b)
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// httpGetJSON issues a simple HTTP GET and decodes a JSON response into out.
// It treats any non-2xx status as an error (returned as httpStatusErr).
func httpGetJSON[T any](u string, out *T) error {
	resp, err := http.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return &httpStatusErr{URL: u, Code: resp.StatusCode}
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// httpPostJSON sends a JSON-encoded POST body and decodes a JSON response into out.
// It treats any non-2xx status as an error (returned as httpStatusErr).
func httpPostJSON[T any](u string, body any, out *T) error {
	b, _ := json.Marshal(body)
	resp, err := http.Post(u, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return &httpStatusErr{URL: u, Code: resp.StatusCode}
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// httpPostAuthJSON is like httpPostJSON but adds a Bearer Authorization header.
// Useful for calling protected APIs with an access token.
func httpPostAuthJSON[T any](u, accessToken string, body any, out *T) error {
	b, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", u, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return &httpStatusErr{URL: u, Code: resp.StatusCode}
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// httpStatusErr represents a non-2xx HTTP response returned by the helpers above.
type httpStatusErr struct {
	URL  string // request URL
	Code int    // HTTP status code
}

func (e *httpStatusErr) Error() string { return "HTTP " + http.StatusText(e.Code) + " for " + e.URL }

// writeJSON writes v as pretty-printed JSON with the correct Content-Type header.
// Errors from Encode are ignored on purpose (best-effort response).
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

// withCORS wraps an http.Handler and adds permissive CORS headers.
// - Allows credentials
// - Mirrors Origin (falls back to issuerBaseURL if no Origin)
// - Handles OPTIONS preflight with 204 No Content
// SECURITY: tighten Allowed-Origin and methods/headers in production.
func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = issuerBaseURL
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// randomString returns an n-length random alphanumeric string.
// Uses crypto/rand for cryptographic randomness, maps bytes into the alphabet.
// Good for nonces, state, PKCE verifiers, etc.
func randomString(n int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	_, _ = rand.Read(b) // best-effort; if this fails, zero bytes remain and are still mapped below
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b)
}
