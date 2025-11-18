// internal/app/zkprequest/service.go
package zkprequest

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"time"

	"zk-wallet-go/internal/app/zkp"

	"github.com/consensys/gnark/backend/groth16"
	"github.com/google/uuid"
)

type Service struct {
	Store    RequestStore
	Circuits CircuitRegistry

	Audience    string
	ResponseURI string
	TTL         time.Duration

	// schema_hash -> serialized VK (server-held)
	vkCache map[string][]byte
	// schema_hash -> serialized PK (server-held)
	pkCache map[string][]byte
	// schema_hash -> canonical JSON schemy
	schemaCache map[string][]byte

	cacheMu sync.RWMutex
}

func NewService(store RequestStore, circuits CircuitRegistry, opts ...func(*Service)) *Service {
	s := &Service{
		Store:       store,
		Circuits:    circuits,
		Audience:    "http://localhost",
		ResponseURI: "http://localhost/v1/presentations/verify",
		TTL:         5 * time.Minute,

		// 🔧 tu były złe typy – teraz wszystko jest []byte
		vkCache:     make(map[string][]byte),
		pkCache:     make(map[string][]byte),
		schemaCache: make(map[string][]byte),
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
// NOTE: w architekturze "Option A" w produkcji weryfikacja i tak będzie po stronie DI,
// więc w wallet-serverze ta funkcja raczej nie będzie wywoływana – ale zostawiamy ją
// poprawioną, żeby kod się kompilował.
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

	// Verify using server-held VK from CircuitRegistry (stary tryb z circuitID)
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
