package zkprequest

type PresentationRequest struct {
	RequestID    string                 `json:"request_id"`
	CircuitID    string                 `json:"circuit_id,omitempty"` // optional (for future registry)
	SchemaJSON   string                 `json:"schema_json"`          // canonical; wallet uses this
	SchemaHash   string                 `json:"schema_hash"`          // sha256 of canonical schema_json
	PublicInputs map[string]interface{} `json:"public_inputs"`
	ResponseURI  string                 `json:"response_uri"`
	ExpiresAt    int64                  `json:"expires_at"`
}

type ProofSubmission struct {
	RequestID    string                 `json:"request_id"`
	ZkpBlobB64   string                 `json:"zkp_blob_b64"`
	PublicInputs map[string]interface{} `json:"public_inputs,omitempty"`
}
