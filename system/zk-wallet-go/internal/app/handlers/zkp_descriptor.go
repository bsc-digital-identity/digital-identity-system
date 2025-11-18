// zkp_descriptor.go
package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"zk-wallet-go/internal/app/config"
)

// PresentationDescriptor mirrors the JSON from /v1/presentations/{id}/descriptor
type PresentationDescriptor struct {
	RequestID string `json:"request_id"`
	Audience  string `json:"audience"`
	ExpiresAt int64  `json:"expires_at"`
	Challenge string `json:"challenge"`
	Nonce     string `json:"nonce"`

	Schema struct {
		Hash string `json:"hash"`
		URI  string `json:"uri"`
	} `json:"schema"`

	Artifacts struct {
		VKURL string `json:"vk_url"`
		PKURL string `json:"pk_url"` // <--- nowość
	} `json:"artifacts"`

	SubmitURL string `json:"submit_url"`
}

type FetchDescriptorRequest struct {
	RequestID string `json:"request_id"`
}

type FetchDescriptorResponse struct {
	Descriptor PresentationDescriptor `json:"descriptor"`
	SchemaJSON json.RawMessage        `json:"schema_json"`
}

// HandleFetchDescriptor:
//   - builds the descriptor URL from request_id and ZkpVerifierBaseURL
//   - fetches the descriptor
//   - fixes any "http://http://" or double-slash bugs in URLs
//   - fetches the schema JSON (server-side, bez CORS problemów)
//   - returns { descriptor, schema_json }
func HandleFetchDescriptor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var in FetchDescriptorRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}

	id := strings.TrimSpace(in.RequestID)
	if id == "" {
		http.Error(w, "missing request_id", http.StatusBadRequest)
		return
	}

	base := strings.TrimRight(config.ZkpVerifierBaseURL, "/")
	descURL := base + "/v1/presentations/" + id + "/descriptor"

	// --- 1) Fetch descriptor JSON ---
	resp, err := http.Get(descURL)
	if err != nil {
		http.Error(w, "descriptor fetch failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		http.Error(
			w,
			fmt.Sprintf("descriptor fetch failed: status %d: %s", resp.StatusCode, string(b)),
			http.StatusBadGateway,
		)
		return
	}

	var descriptor PresentationDescriptor
	if err := json.NewDecoder(resp.Body).Decode(&descriptor); err != nil {
		http.Error(w, "descriptor json error: "+err.Error(), http.StatusBadRequest)
		return
	}

	// --- 2) Naprawiamy krzaczaste URL-e z DI ---
	descriptor.Schema.URI = normalizeWeirdURL(descriptor.Schema.URI)
	descriptor.SubmitURL = normalizeWeirdURL(descriptor.SubmitURL)
	descriptor.Artifacts.VKURL = normalizeWeirdURL(descriptor.Artifacts.VKURL)
	descriptor.Audience = normalizeWeirdURL(descriptor.Audience)

	if strings.TrimSpace(descriptor.Schema.URI) == "" {
		http.Error(w, "descriptor has empty schema.uri", http.StatusBadRequest)
		return
	}

	// --- 3) Fetch schema JSON z poprawionego URI ---
	schemaResp, err := http.Get(descriptor.Schema.URI)
	if err != nil {
		http.Error(w, "schema fetch failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer schemaResp.Body.Close()

	if schemaResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(schemaResp.Body)
		http.Error(
			w,
			fmt.Sprintf("schema fetch failed: status %d: %s", schemaResp.StatusCode, string(b)),
			http.StatusBadGateway,
		)
		return
	}

	schemaBytes, err := io.ReadAll(schemaResp.Body)
	if err != nil {
		http.Error(w, "schema read failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	out := FetchDescriptorResponse{
		Descriptor: descriptor,
		SchemaJSON: json.RawMessage(schemaBytes),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// normalizeWeirdURL usuwa duplikaty "http://http://" i nadmiarowe slashe.
func normalizeWeirdURL(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return u
	}

	// napraw typowe "http://http://"
	for strings.Contains(u, "http://http://") {
		u = strings.ReplaceAll(u, "http://http://", "http://")
	}
	for strings.Contains(u, "https://http://") {
		u = strings.ReplaceAll(u, "https://http://", "https://")
	}

	// proste zbijanie nadmiarowych "//" (poza "://")
	parts := strings.SplitN(u, "://", 2)
	if len(parts) == 2 {
		scheme, rest := parts[0], parts[1]
		// zamień "//" na "/" w części po schemacie
		for strings.Contains(rest, "//") {
			rest = strings.ReplaceAll(rest, "//", "/")
		}
		u = scheme + "://" + rest
	}

	return u
}

type VerifyProxyRequest struct {
	RequestID    string         `json:"request_id"`
	ZkpBlobB64   string         `json:"zkp_blob_b64"`
	PublicInputs map[string]any `json:"public_inputs,omitempty"`
	Challenge    string         `json:"challenge,omitempty"`
}

// POST /wallet/zkp/verify
// Body: { "request_id": "...", "zkp_blob_b64": "...", ... }
// Proxy do {ZKP_VERIFIER_BASE_URL}/v1/presentations/verify
func HandleVerifyProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var in VerifyProxyRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}

	id := strings.TrimSpace(in.RequestID)
	if id == "" {
		http.Error(w, "missing request_id", http.StatusBadRequest)
		return
	}
	blob := strings.TrimSpace(in.ZkpBlobB64)
	if blob == "" {
		http.Error(w, "missing zkp_blob_b64", http.StatusBadRequest)
		return
	}

	base := strings.TrimRight(config.ZkpVerifierBaseURL, "/")
	target := base + "/v1/presentations/verify"

	// budujemy payload zgodny z VerifyIn
	payload := map[string]any{
		"request_id":   id,
		"zkp_blob_b64": blob,
	}
	if len(in.PublicInputs) > 0 {
		payload["public_inputs"] = in.PublicInputs
	}
	if strings.TrimSpace(in.Challenge) != "" {
		payload["challenge"] = strings.TrimSpace(in.Challenge)
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "cannot marshal payload: "+err.Error(), http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest(http.MethodPost, target, bytes.NewReader(bodyBytes))
	if err != nil {
		http.Error(w, "cannot create request: "+err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "verify proxy failed: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
