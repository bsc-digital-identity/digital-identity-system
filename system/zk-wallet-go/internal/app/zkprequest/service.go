package zkprequest

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"time"
	"zk-wallet-go/internal/app/zkp"

	"github.com/consensys/gnark/backend/groth16"
	"github.com/google/uuid"
)

type Service struct {
	Store    RequestStore
	Circuits CircuitRegistry

	// Config
	Audience    string        // e.g., https://your-verifier.example
	ResponseURI string        // e.g., https://your-verifier.example/zkp/verify
	TTL         time.Duration // validity for requests
	// Security toggle: if true, use server-held VK by circuit; if false, accept client VK (dev only)
	VerifyWithServerVK bool
}

func (s *Service) defaultTTL() time.Duration {
	if s.TTL <= 0 {
		return 5 * time.Minute
	}
	return s.TTL
}

// Build public inputs that must be bound into the proof (nonce, aud, current date).
func (s *Service) defaultPublicInputs(now time.Time) map[string]interface{} {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return map[string]interface{}{
		"current_year":  now.Year(),
		"current_month": int(now.Month()),
		"current_day":   now.Day(),
		"aud":           s.Audience,
		"nonce":         uuid.NewString(),
	}
}

func (s *Service) CreateRequest(circuitID string, includeSchema bool) (PresentationRequest, error) {
	if s.Store == nil || s.Circuits == nil {
		return PresentationRequest{}, errors.New("service not initialized")
	}
	req := PresentationRequest{
		RequestID:    uuid.NewString(),
		CircuitID:    circuitID,
		PublicInputs: s.defaultPublicInputs(time.Now().UTC()),
		ResponseURI:  s.ResponseURI,
		ExpiresAt:    time.Now().UTC().Add(s.defaultTTL()).Unix(),
	}
	if includeSchema {
		if schema, ok := s.Circuits.SchemaJSON(circuitID); ok {
			req.SchemaJSON = schema
		}
	}
	if err := s.Store.Save(req); err != nil {
		return PresentationRequest{}, err
	}
	return req, nil
}

// VerifySubmission checks the proof against the request.
// Returns the parsed request (for audit) on success.
func (s *Service) VerifySubmission(sub ProofSubmission) (PresentationRequest, error) {
	if sub.RequestID == "" || sub.ZkpBlobB64 == "" {
		return PresentationRequest{}, errors.New("missing request_id or zkp_blob_b64")
	}
	req, ok := s.Store.Load(sub.RequestID)
	if !ok {
		return PresentationRequest{}, errors.New("request not found or already used")
	}
	if time.Now().Unix() > req.ExpiresAt {
		s.Store.Delete(sub.RequestID)
		return PresentationRequest{}, errors.New("request expired")
	}

	// Reconstruct proof package
	blob, err := base64.StdEncoding.DecodeString(sub.ZkpBlobB64)
	if err != nil {
		return PresentationRequest{}, errors.New("invalid zkp_blob_b64")
	}
	res, err := zkp.ReconstructZkpResult(blob)
	if err != nil {
		return PresentationRequest{}, errors.New("invalid proof package")
	}

	// (Optional) double-check public inputs out-of-band
	if len(sub.PublicInputs) > 0 {
		// Basic JSON equality check
		want, _ := json.Marshal(req.PublicInputs)
		got, _ := json.Marshal(sub.PublicInputs)
		if string(want) != string(got) {
			return PresentationRequest{}, errors.New("public inputs mismatch")
		}
	}

	// Verify
	if s.VerifyWithServerVK {
		// Use server-held VK for this circuit id (prod path)
		vkBytes, ok := s.Circuits.VerifyingKey(req.CircuitID)
		if !ok {
			return PresentationRequest{}, errors.New("server VK not found for circuit")
		}
		vk := groth16.NewVerifyingKey(zkp.ElipticalCurveID) // keep same curve
		if _, err := vk.ReadFrom(bytesReader(vkBytes)); err != nil {
			return PresentationRequest{}, errors.New("cannot load server VK")
		}
		if err := groth16.Verify(res.Proof, vk, res.PublicWitness); err != nil {
			return PresentationRequest{}, errors.New("verify failed")
		}
	} else {
		// DEV ONLY: verify with client-provided VK from the blob
		if err := groth16.Verify(res.Proof, res.VerifyingKey, res.PublicWitness); err != nil {
			return PresentationRequest{}, errors.New("verify failed")
		}
	}

	// one-shot consume
	s.Store.Delete(sub.RequestID)
	return req, nil
}

// helper: []byte reader without importing bytes in caller
func bytesReader(b []byte) *bytesReaderImpl { return &bytesReaderImpl{b: b} }

type bytesReaderImpl struct {
	b []byte
	i int
}

func (r *bytesReaderImpl) Read(p []byte) (int, error) {
	if r.i >= len(r.b) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.i:])
	r.i += n
	return n, nil
}
