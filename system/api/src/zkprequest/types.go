// file: internal/app/zkprequest/model.go
package zkprequest

// ---- Requests coming FROM wallet → verifier ----

// ---- Requests coming FROM wallet → verifier ----
type ProofSubmission struct {
	RequestID    string         `json:"request_id"`
	ZkpBlobB64   string         `json:"zkp_blob_b64"`
	PublicInputs map[string]any `json:"public_inputs,omitempty"`
	Challenge    string         `json:"challenge,omitempty"` // <--- NEW (lustro z VerifyIn)
}

// ---- Objects your verifier issues TO the wallet (or RP) ----
type PresentationRequest struct {
	RequestID    string         `json:"request_id"`
	SchemaJSON   string         `json:"schema_json"`
	SchemaHash   string         `json:"schema_hash"`
	PublicInputs map[string]any `json:"public_inputs"`
	ResponseURI  string         `json:"response_uri"`
	ExpiresAt    int64          `json:"expires_at"`

	// server-only
	CallbackURL    string `json:"-"`
	CallbackSecret string `json:"-"`
}

// ---- Inputs/Outputs for public API ----
type CreatePresentationIn struct {
	SchemaJSON     string         `json:"schema_json" binding:"required"`
	PublicInputs   map[string]any `json:"public_inputs,omitempty"`
	ExpiresIn      int64          `json:"expires_in,omitempty"`
	CallbackURL    string         `json:"callback_url,omitempty"`
	CallbackSecret string         `json:"callback_secret,omitempty"`
}

type CreatePresentationOut struct {
	Request     PresentationRequest `json:"request"`
	RequestURL  string              `json:"request_url"`
	DeepLink    string              `json:"deeplink"`
	QRPngBase64 string              `json:"qr_png_b64"`
}

// ---- Verify ----
type VerifyIn struct {
	RequestID    string         `json:"request_id" binding:"required"`
	ZkpBlobB64   string         `json:"zkp_blob_b64" binding:"required"`
	PublicInputs map[string]any `json:"public_inputs,omitempty"`
	Challenge    string         `json:"challenge,omitempty"` // <--- NEW
}

type VerifyOut struct {
	OK         bool   `json:"ok"`
	VerifiedAt string `json:"verified_at"`
}

type StatusOut struct {
	State string `json:"state"`
}
