package zkprequest

// What the verifier serves (usually via QR/deeplink) to the wallet.
type PresentationRequest struct {
	RequestID    string                 `json:"request_id"`
	CircuitID    string                 `json:"circuit_id"`
	SchemaJSON   string                 `json:"schema_json"`   // optional if you only pass circuit id
	PublicInputs map[string]interface{} `json:"public_inputs"` // e.g., current date, aud, nonce
	ResponseURI  string                 `json:"response_uri"`
	ExpiresAt    int64                  `json:"expires_at"` // unix seconds
}

// What the wallet posts back to the verifier.
type ProofSubmission struct {
	RequestID    string                 `json:"request_id"`
	ZkpBlobB64   string                 `json:"zkp_blob_b64"`
	PublicInputs map[string]interface{} `json:"public_inputs,omitempty"` // optional echo
}
