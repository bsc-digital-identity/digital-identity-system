package zkprequest

import (
	"encoding/json"
	"net/http"
	"pkg-common/utilities"
	"time"
	"zk-wallet-go/pkg/util"
)

type CreatePresentationIn struct {
	SchemaJSON   string         `json:"schema_json"`
	PublicInputs map[string]any `json:"public_inputs,omitempty"` // (optional; currently unused override)
	ExpiresIn    int64          `json:"expires_in,omitempty"`    // (optional; default 300s; you can use it later)
}

var ZKPService *Service // set in main

// POST /presentations/create
func HandleCreatePresentation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var in CreatePresentationIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if in.SchemaJSON == "" {
		http.Error(w, "schema_json required", http.StatusBadRequest)
		return
	}
	req, err := ZKPService.CreateRequestFromSchema(in.SchemaJSON, time.Now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	utilities.WriteJson(w, req)
}

type VerifyIn struct {
	RequestID    string         `json:"request_id"`
	ZkpBlobB64   string         `json:"zkp_blob_b64"`
	PublicInputs map[string]any `json:"public_inputs,omitempty"`
}

// POST /presentations/verify
func HandleVerifyPresentation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var in VerifyIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	req, err := ZKPService.VerifySubmission(ProofSubmission{
		RequestID:    in.RequestID,
		ZkpBlobB64:   in.ZkpBlobB64,
		PublicInputs: in.PublicInputs,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	util.WriteJSON(w, map[string]any{
		"ok":          true,
		"request_id":  req.RequestID,
		"verified_at": time.Now().UTC().Format(time.RFC3339),
	})
}
