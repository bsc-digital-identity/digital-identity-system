// service.go
package zkprequest

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/consensys/gnark/backend/groth16"
	"github.com/google/uuid"

	"pkg-common/zkp"
)

// --- new: verdict struct kept short-term in memory (no DB needed) ---
type verdict struct {
	OK         bool
	Reason     string
	State      string    // "verified" | "failed" | "expired"
	VerifiedAt time.Time // set only on success
}

type Service struct {
	Store    RequestStore
	Circuits CircuitRegistry // zostawione pod przyszły registry/circuitID mode

	Audience    string
	ResponseURI string
	TTL         time.Duration

	// schema_hash -> serialized VK (server-held)
	vkCache map[string][]byte

	// NEW: schema_hash -> serialized PK (server-held, dla walleta)
	pkCache map[string][]byte

	// NEW: schema_hash -> canonical schema JSON (raw bytes)
	schemaCache map[string][]byte

	// allow accepting ad-hoc schema_json from RP
	AllowAdHocSchema bool

	// --- blocking waiters ---
	waiters   map[string][]chan verifyResult // requestID -> list of listeners
	waitersMu sync.Mutex

	// --- new: short-lived verdict memory ---
	verdicts   map[string]verdict
	verdictsMu sync.Mutex

	// wspólny mutex dla vkCache + pkCache + schemaCache
	cacheMu sync.RWMutex
}

type verifyResult struct {
	OK     bool
	Reason string
}

func NewService(store RequestStore, opts ...func(*Service)) *Service {
	s := &Service{
		Store:            store,
		Audience:         "http://localhost",
		ResponseURI:      "http://localhost/presentations/verify",
		TTL:              5 * time.Minute,
		vkCache:          make(map[string][]byte),
		pkCache:          make(map[string][]byte),
		schemaCache:      make(map[string][]byte),
		AllowAdHocSchema: true,
		waiters:          make(map[string][]chan verifyResult),
		verdicts:         make(map[string]verdict),
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

	// (opcjonalnie) walidacja schematu/operators/rozmiarów

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
	return req, nil
}

// VerifySubmission reconstructs the ZKP blob, loads the server-held VK by schema_hash,
// verifies the proof, and consumes the request. Powiadamia waiters + webhook + zapisuje werdykt.
// VerifySubmission reconstructs the ZKP blob, loads the server-held VK by schema_hash,
// verifies the proof, and consumes the request. Powiadamia waiters + webhook + zapisuje werdykt.
func (s *Service) VerifySubmission(sub ProofSubmission) (PresentationRequest, error) {
	// small helper to record final verdict + notify
	recordFail := func(req *PresentationRequest, reason string) (PresentationRequest, error) {
		if req != nil {
			// set final verdict (kept for a short time)
			s.setVerdict(req.RequestID, verdict{
				OK:     false,
				Reason: reason,
				State:  "failed",
			})
			// notify dependents
			s.postWebhook(*req, false, reason)
			s.notifyWaiters(req.RequestID, verifyResult{OK: false, Reason: reason})
		}
		return PresentationRequest{}, errors.New(reason)
	}

	if sub.RequestID == "" || sub.ZkpBlobB64 == "" {
		return PresentationRequest{}, errors.New("missing request_id or zkp_blob_b64")
	}
	req, ok := s.Store.Load(sub.RequestID)
	if !ok {
		return PresentationRequest{}, errors.New("request not found or already used")
	}
	now := time.Now().Unix()
	if now > req.ExpiresAt {
		s.Store.Delete(sub.RequestID)
		return recordFail(&req, "request expired")
	}

	// 🔐 Echo-check tylko aud + nonce (binding + anti-replay)
	if len(sub.PublicInputs) > 0 {
		// aud musi się zgadzać z Audience serwera
		wantAud := fmt.Sprint(s.Audience)
		gotAud := fmt.Sprint(sub.PublicInputs["aud"])
		if wantAud != gotAud {
			return recordFail(&req,
				fmt.Sprintf("aud mismatch: want=%s got=%s", wantAud, gotAud),
			)
		}

		// nonce musi się zgadzać z tym, co DI wygenerował przy create
		wantNonce := fmt.Sprint(req.PublicInputs["nonce"])
		gotNonce := fmt.Sprint(sub.PublicInputs["nonce"])
		if wantNonce == "" || gotNonce == "" || wantNonce != gotNonce {
			return recordFail(&req,
				fmt.Sprintf("nonce mismatch: want=%s got=%s", wantNonce, gotNonce),
			)
		}
	}

	raw, err := base64.StdEncoding.DecodeString(sub.ZkpBlobB64)
	if err != nil {
		return recordFail(&req, "invalid zkp_blob_b64")
	}
	pkg, err := zkp.ReconstructZkpResult(raw)
	if err != nil {
		return recordFail(&req, "invalid proof package")
	}

	// Verify with server-held VK (never trust client VK in pkg)
	s.cacheMu.RLock()
	vkb := s.vkCache[req.SchemaHash]
	s.cacheMu.RUnlock()
	if len(vkb) == 0 {
		return recordFail(&req, "server VK not found")
	}
	vk := groth16.NewVerifyingKey(zkp.ElipticalCurveID)
	if _, err := vk.ReadFrom(newBytesReader(vkb)); err != nil {
		return recordFail(&req, "cannot load server VK")
	}
	if err := groth16.Verify(pkg.Proof, vk, pkg.PublicWitness); err != nil {
		return recordFail(&req, "verify failed")
	}

	// success: consume, record verdict, webhook, notify
	s.Store.Delete(sub.RequestID)
	verifiedAt := time.Now().UTC()

	s.setVerdict(req.RequestID, verdict{
		OK:         true,
		State:      "verified",
		VerifiedAt: verifiedAt,
	})

	s.postWebhook(req, true, "")
	s.notifyWaiters(req.RequestID, verifyResult{OK: true})

	return req, nil
}

// --- NEW ---
// MockVerify marks a request as verified (DEV ONLY). No proof required.
func (s *Service) MockVerify(requestID string) (PresentationRequest, error) {
	req, ok := s.Store.Load(requestID)
	if !ok {
		return PresentationRequest{}, errors.New("request not found or already used")
	}
	now := time.Now().UTC()
	if now.Unix() > req.ExpiresAt {
		s.Store.Delete(requestID)
		s.setVerdict(requestID, verdict{OK: false, State: "expired", Reason: "request expired"})
		s.postWebhook(req, false, "request expired")
		s.notifyWaiters(requestID, verifyResult{OK: false, Reason: "request expired"})
		return PresentationRequest{}, errors.New("request expired")
	}

	// consume & set final verdict
	s.Store.Delete(requestID)
	s.setVerdict(requestID, verdict{OK: true, State: "verified", VerifiedAt: now})
	s.postWebhook(req, true, "")
	s.notifyWaiters(requestID, verifyResult{OK: true})
	return req, nil
}

// ---- blocking wait ---- (unchanged)

func (s *Service) WaitForResult(requestID string, timeout time.Duration) (verifyResult, bool) {
	ch := make(chan verifyResult, 1)

	s.waitersMu.Lock()
	s.waiters[requestID] = append(s.waiters[requestID], ch)
	s.waitersMu.Unlock()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case res := <-ch:
		return res, true
	case <-timer.C:
		// cleanup
		s.waitersMu.Lock()
		list := s.waiters[requestID]
		newList := make([]chan verifyResult, 0, len(list))
		for _, c := range list {
			if c != ch {
				newList = append(newList, c)
			}
		}
		if len(newList) == 0 {
			delete(s.waiters, requestID)
		} else {
			s.waiters[requestID] = newList
		}
		s.waitersMu.Unlock()
		return verifyResult{OK: false, Reason: "timeout"}, false
	}
}

func (s *Service) notifyWaiters(requestID string, res verifyResult) {
	s.waitersMu.Lock()
	list := s.waiters[requestID]
	delete(s.waiters, requestID)
	s.waitersMu.Unlock()

	for _, ch := range list {
		select {
		case ch <- res:
		default:
		}
	}
}

// ---- webhook ---- (unchanged with corrected state)

func (s *Service) postWebhook(req PresentationRequest, ok bool, reason string) {
	if req.CallbackURL == "" {
		return
	}
	state := "verified"
	if !ok {
		state = "failed"
	}
	payload := map[string]any{
		"request_id":  req.RequestID,
		"schema_hash": req.SchemaHash,
		"state":       state,
		"ok":          ok,
		"reason":      reason,
		"verified_at": time.Now().UTC().Format(time.RFC3339),
	}
	b, _ := json.Marshal(payload)

	r, _ := http.NewRequest(http.MethodPost, req.CallbackURL, bytes.NewReader(b))
	r.Header.Set("Content-Type", "application/json")

	if req.CallbackSecret != "" {
		mac := hmac.New(sha256.New, []byte(req.CallbackSecret))
		mac.Write(b)
		sig := hex.EncodeToString(mac.Sum(nil))
		r.Header.Set("X-ZKP-Signature", sig)
	}

	http.DefaultClient.Do(r) //nolint:errcheck
}

// ---- helpers ----

func (s *Service) setVerdict(id string, v verdict) {
	s.verdictsMu.Lock()
	s.verdicts[id] = v
	s.verdictsMu.Unlock()

	go func() {
		time.Sleep(15 * time.Minute)
		s.verdictsMu.Lock()
		delete(s.verdicts, id)
		s.verdictsMu.Unlock()
	}()
}

func (s *Service) getVerdict(id string) (verdict, bool) {
	s.verdictsMu.Lock()
	v, ok := s.verdicts[id]
	s.verdictsMu.Unlock()
	return v, ok
}

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
	if len(out) > 0 && out[len(out)-1] == '\n' {
		out = out[:len(out)-1]
	}
	return out, nil
}

func newBytesReader(b []byte) *bytesReader { return &bytesReader{b: b} }

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

func (s *Service) LoadRequest(id string) (PresentationRequest, bool) {
	return s.Store.Load(id)
}
