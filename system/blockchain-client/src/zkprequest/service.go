package zkprequest

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"io"
	"log"
	"time"

	"github.com/consensys/gnark/backend/groth16"
	"github.com/google/uuid"

	// Adjust this import to your module path:
	"pkg-common/zkp"
)

// Service wires the presentation (proof) request lifecycle.
type Service struct {
	Store    RequestStore
	Circuits CircuitRegistry // kept for future registry/circuitID mode

	Audience    string
	ResponseURI string
	TTL         time.Duration

	// schema_hash -> serialized VK (server-held)
	vkCache map[string][]byte

	// allow accepting ad-hoc schema_json from RP
	AllowAdHocSchema bool
}

func NewService(store RequestStore, opts ...func(*Service)) *Service {
	s := &Service{
		Store:            store,
		Audience:         "http://localhost",
		ResponseURI:      "http://localhost/presentations/verify",
		TTL:              5 * time.Minute,
		vkCache:          make(map[string][]byte),
		AllowAdHocSchema: true,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

func (s *Service) defaultTTL() time.Duration {
	if s.TTL <= 0 {
		return 5 * time.Minute
	}
	return s.TTL
}

func (s *Service) defaultPublicInputs(now time.Time) map[string]any {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return map[string]any{
		"current_year":  now.Year(),
		"current_month": int(now.Month()),
		"current_day":   now.Day(),
		"aud":           s.Audience,
		"nonce":         uuid.NewString(),
	}
}

// CreateRequestFromSchema accepts a flexible DynamicCircuit schema JSON,
// canonicalizes + hashes it, ensures VK is cached, and returns a one-shot request.
func (s *Service) CreateRequestFromSchema(schemaJSON string, now time.Time) (PresentationRequest, error) {
	if s.Store == nil {
		return PresentationRequest{}, errors.New("service not initialized")
	}
	if !s.AllowAdHocSchema {
		return PresentationRequest{}, errors.New("ad-hoc schema disabled")
	}
	canon, err := canonicalJSON(schemaJSON)
	if err != nil {
		return PresentationRequest{}, fmt.Errorf("invalid schema_json: %w", err)
	}

	// (Optional) add a guard here to whitelist operators / size limits.
	// if err := validateDynamicSchema([]byte(canon)); err != nil { ... }

	schemaHash, err := s.ensureVKForSchema(canon)
	if err != nil {
		return PresentationRequest{}, fmt.Errorf("setup failed: %w", err)
	}

	req := PresentationRequest{
		RequestID:    uuid.NewString(),
		SchemaJSON:   canon,
		SchemaHash:   schemaHash,
		PublicInputs: s.defaultPublicInputs(now),
		ResponseURI:  s.ResponseURI,
		ExpiresAt:    now.UTC().Add(s.defaultTTL()).Unix(),
	}
	if err := s.Store.Save(req); err != nil {
		return PresentationRequest{}, err
	}
	log.Printf("[zkp] CreateRequestFromSchema ok: request_id=%s schema_hash=%s", req.RequestID, req.SchemaHash)
	return req, nil
}

// VerifySubmission reconstructs the ZKP blob, loads the server-held VK by schema_hash,
// verifies the proof, and consumes the request.
func (s *Service) VerifySubmission(sub ProofSubmission) (PresentationRequest, error) {
	log.Printf("[zkp] VerifySubmission start: request_id=%s", sub.RequestID)

	if sub.RequestID == "" || sub.ZkpBlobB64 == "" {
		log.Printf("[zkp] VerifySubmission error: missing request_id or zkp_blob_b64")
		return PresentationRequest{}, errors.New("missing request_id or zkp_blob_b64")
	}

	req, ok := s.Store.Load(sub.RequestID)
	if !ok {
		log.Printf("[zkp] VerifySubmission error: request not found or already used: %s", sub.RequestID)
		return PresentationRequest{}, errors.New("request not found or already used")
	}
	if time.Now().Unix() > req.ExpiresAt {
		log.Printf("[zkp] VerifySubmission error: request expired: %s", sub.RequestID)
		s.Store.Delete(sub.RequestID)
		return PresentationRequest{}, errors.New("request expired")
	}

	// 🔒 Aud / nonce binding (zamiast pełnego porównania JSON-ów)
	if len(sub.PublicInputs) > 0 {
		wantAud := fmt.Sprint(s.Audience)
		gotAud := fmt.Sprint(sub.PublicInputs["aud"])
		log.Printf("[zkp] VerifySubmission aud check: want=%q got=%q", wantAud, gotAud)
		if wantAud != gotAud {
			log.Printf("[zkp] VerifySubmission error: aud mismatch")
			return PresentationRequest{}, errors.New("aud mismatch")
		}

		wantNonce := fmt.Sprint(req.PublicInputs["nonce"])
		gotNonce := fmt.Sprint(sub.PublicInputs["nonce"])
		log.Printf("[zkp] VerifySubmission nonce check: want=%q got=%q", wantNonce, gotNonce)
		if wantNonce == "" || gotNonce == "" || wantNonce != gotNonce {
			log.Printf("[zkp] VerifySubmission error: nonce mismatch")
			return PresentationRequest{}, errors.New("nonce mismatch")
		}
	} else {
		log.Printf("[zkp] VerifySubmission warning: no publicInputs provided in submission")
	}

	raw, err := base64.StdEncoding.DecodeString(sub.ZkpBlobB64)
	if err != nil {
		log.Printf("[zkp] VerifySubmission error: invalid zkp_blob_b64: %v", err)
		return PresentationRequest{}, errors.New("invalid zkp_blob_b64")
	}
	pkg, err := zkp.ReconstructZkpResult(raw)
	if err != nil {
		log.Printf("[zkp] VerifySubmission error: invalid proof package: %v", err)
		return PresentationRequest{}, errors.New("invalid proof package")
	}

	// Verify with server-held VK (never trust client VK in pkg)
	vkb := s.vkCache[req.SchemaHash]
	if len(vkb) == 0 {
		log.Printf("[zkp] VerifySubmission error: server VK not found for schema_hash=%s", req.SchemaHash)
		return PresentationRequest{}, errors.New("server VK not found")
	}
	vk := groth16.NewVerifyingKey(ecc.ID(zkp.ElipticalCurveID))
	if _, err := vk.ReadFrom(newBytesReader(vkb)); err != nil {
		log.Printf("[zkp] VerifySubmission error: cannot load server VK: %v", err)
		return PresentationRequest{}, errors.New("cannot load server VK")
	}
	if err := groth16.Verify(pkg.Proof, vk, pkg.PublicWitness); err != nil {
		log.Printf("[zkp] VerifySubmission error: verify failed: %v", err)
		return PresentationRequest{}, errors.New("verify failed")
	}

	// one-shot consume
	s.Store.Delete(sub.RequestID)
	log.Printf("[zkp] VerifySubmission ok: request_id=%s", sub.RequestID)
	return req, nil
}

// ---- helpers ----

// canonicalJSON re-encodes JSON with stable formatting.
// (If you need strict key sorting, implement recursive map key sorting.)
func canonicalJSON(raw string) (string, error) {
	var v any
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return "", err
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return "", err
	}
	out := buf.String()
	// json.Encoder adds a trailing newline
	if len(out) > 0 && out[len(out)-1] == '\n' {
		out = out[:len(out)-1]
	}
	return out, nil
}

// tiny io.Reader over []byte (to avoid importing bytes.NewReader everywhere)
func newBytesReader(b []byte) *bytesReader {
	return &bytesReader{b: b}
}

type bytesReader struct {
	b []byte
	i int
}

func (r *bytesReader) Read(p []byte) (int, error) {
	if r.i >= len(r.b) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.i:])
	r.i += n
	return n, nil
}
