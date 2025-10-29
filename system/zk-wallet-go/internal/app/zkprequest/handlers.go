package zkprequest

import (
	"encoding/json"
	"net/http"
	"time"
)

// Handlers exposes ready-to-mount http.Handlers that call Service methods.
type Handlers struct {
	Svc *Service
}

func (h *Handlers) Request(w http.ResponseWriter, r *http.Request) {
	// For demo we hardcode circuit id; you can also read from query (?circuit_id=...)
	circuitID := "age_over_18@1"
	includeSchema := true

	req, err := h.Svc.CreateRequest(circuitID, includeSchema)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(req)
}

func (h *Handlers) Verify(w http.ResponseWriter, r *http.Request) {
	var sub ProofSubmission
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	req, err := h.Svc.VerifySubmission(sub)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":          true,
		"request_id":  sub.RequestID,
		"verified_at": time.Now().UTC(),
		"circuit_id":  req.CircuitID,
	})
}
