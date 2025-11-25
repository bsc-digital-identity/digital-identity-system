// zkp_prove.go
package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"strings"

	"zk-wallet-go/internal/app/config"
	"zk-wallet-go/internal/app/vcstore"
	"zk-wallet-go/internal/app/zkp"
	"zk-wallet-go/pkg/util"

	"github.com/lestrrat-go/jwx/v2/jws"
)

// ----------------------------
//   Handler z VCStore
// ----------------------------

type ZkpHandler struct {
	VCs vcstore.VCStore
}

func NewZkpHandler(store vcstore.VCStore) *ZkpHandler {
	return &ZkpHandler{VCs: store}
}

// helper: pobiera wartość po ścieżce w stylu "vc.credentialSubject.birth_ts"
func getByPath(m map[string]any, path string) (any, bool) {
	if m == nil {
		return nil, false
	}
	parts := strings.Split(path, ".")
	var cur any = m
	for _, p := range parts {
		asMap, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		v, ok := asMap[p]
		if !ok {
			return nil, false
		}
		cur = v
	}
	return cur, true
}

// helper: typowy case VCDM -> vc.credentialSubject
func extractCredentialSubject(claims map[string]any) map[string]any {
	rawVC, ok := claims["vc"]
	if !ok {
		return nil
	}
	vcMap, ok := rawVC.(map[string]any)
	if !ok {
		return nil
	}
	rawCS, ok := vcMap["credentialSubject"]
	if !ok {
		return nil
	}
	csMap, ok := rawCS.(map[string]any)
	if !ok {
		return nil
	}
	return csMap
}

// fetchDescriptorAndSchema pobiera descriptor + schema JSON + PK dla danego request_id.
func fetchDescriptorSchemaAndPK(requestID string) (PresentationDescriptor, []byte, []byte, error) {
	var desc PresentationDescriptor

	base := strings.TrimRight(config.ZkpVerifierBaseURL, "/")
	descURL := base + "/v1/presentations/" + requestID + "/descriptor"

	log.Printf("[zkp] fetching descriptor for request_id=%s from %s", requestID, descURL)

	resp, err := http.Get(descURL)
	if err != nil {
		return desc, nil, nil, fmt.Errorf("descriptor fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return desc, nil, nil, fmt.Errorf("descriptor fetch: status %d: %s", resp.StatusCode, string(b))
	}

	if err := json.NewDecoder(resp.Body).Decode(&desc); err != nil {
		return desc, nil, nil, fmt.Errorf("descriptor json: %w", err)
	}

	// normalizacja URL-i
	desc.Schema.URI = normalizeWeirdURL(desc.Schema.URI)
	desc.SubmitURL = normalizeWeirdURL(desc.SubmitURL)
	desc.Artifacts.VKURL = normalizeWeirdURL(desc.Artifacts.VKURL)
	desc.Artifacts.PKURL = normalizeWeirdURL(desc.Artifacts.PKURL)
	desc.Audience = normalizeWeirdURL(desc.Audience)

	log.Printf("[zkp] descriptor loaded: schema_uri=%s pk_url=%s submit_url=%s audience=%s",
		desc.Schema.URI, desc.Artifacts.PKURL, desc.SubmitURL, desc.Audience)

	if strings.TrimSpace(desc.Schema.URI) == "" {
		return desc, nil, nil, fmt.Errorf("descriptor has empty schema.uri")
	}
	if strings.TrimSpace(desc.Artifacts.PKURL) == "" {
		return desc, nil, nil, fmt.Errorf("descriptor has empty artifacts.pk_url")
	}

	// fetch schema
	log.Printf("[zkp] fetching schema from %s", desc.Schema.URI)
	schemaResp, err := http.Get(desc.Schema.URI)
	if err != nil {
		return desc, nil, nil, fmt.Errorf("schema fetch: %w", err)
	}
	defer schemaResp.Body.Close()

	if schemaResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(schemaResp.Body)
		return desc, nil, nil, fmt.Errorf("schema fetch: status %d: %s", schemaResp.StatusCode, string(b))
	}

	schemaBytes, err := io.ReadAll(schemaResp.Body)
	if err != nil {
		return desc, nil, nil, fmt.Errorf("schema read: %w", err)
	}

	// fetch PK
	log.Printf("[zkp] fetching PK from %s", desc.Artifacts.PKURL)
	pkResp, err := http.Get(desc.Artifacts.PKURL)
	if err != nil {
		return desc, nil, nil, fmt.Errorf("pk fetch: %w", err)
	}
	defer pkResp.Body.Close()

	if pkResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(pkResp.Body)
		return desc, nil, nil, fmt.Errorf("pk fetch: status %d: %s", pkResp.StatusCode, string(b))
	}

	pkBytes, err := io.ReadAll(pkResp.Body)
	if err != nil {
		return desc, nil, nil, fmt.Errorf("pk read: %w", err)
	}

	log.Printf("[zkp] descriptor/schema/pk fetch OK for request_id=%s", requestID)
	return desc, schemaBytes, pkBytes, nil
}

// ----------------------------
//   API: /wallet/zkp/create
// ----------------------------

type ZkpCreateRequest struct {
	RequestID string                 `json:"request_id"`
	Inputs    map[string]interface{} `json:"inputs,omitempty"`
}

type ZkpCreateResponse struct {
	RequestID        string `json:"request_id"`
	SchemaHash       string `json:"schema_hash"`
	SchemaURI        string `json:"schema_uri"`
	Audience         string `json:"audience"`
	ExpiresAt        int64  `json:"expires_at"`
	SubmitURL        string `json:"submit_url"`
	ZkpBorshBase64   string `json:"zkp_borsh_base64"`
	ProofLength      int    `json:"proof_length"`
	PublicWitnessOK  bool   `json:"public_witness_ok"`
	VerifierResponse any    `json:"verifier_response,omitempty"`
}

// HandleZkpCreate:
//   - przyjmuje { request_id }
//   - pobiera descriptor + schema + PK z DI
//   - buduje assignments WYŁĄCZNIE z VCStore,
//     dopasowując nazwy pól 1:1 do schema.fields[*].name
//   - dokleja aud / nonce z descriptor, jeśli są w schemie
//   - generuje dowód Groth16 (ProveDynamicFromSchema)
//   - serializuje ZkpResult do Borsh + base64
//   - buduje public_inputs z PUBLIC fields
//   - wywołuje DI /v1/presentations/verify
//   - zwraca ZkpCreateResponse
func (h *ZkpHandler) HandleZkpCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var in ZkpCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}
	reqID := strings.TrimSpace(in.RequestID)
	if reqID == "" {
		http.Error(w, "missing request_id", http.StatusBadRequest)
		return
	}

	log.Printf("[zkp] HandleZkpCreate start, request_id=%s", reqID)

	// 1) descriptor + schema + pk
	desc, schemaJSON, pkBytes, err := fetchDescriptorSchemaAndPK(reqID)
	if err != nil {
		log.Printf("[zkp] descriptor/schema/pk fetch failed: %v", err)
		http.Error(w, "descriptor/schema/pk fetch failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 2) parse schema
	schemaDef, err := zkp.ParseSchema(schemaJSON)
	if err != nil {
		log.Printf("[zkp] schema parse failed: %v", err)
		http.Error(w, "schema parse failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[zkp] schema parsed: hash=%s, fields=%d", desc.Schema.Hash, len(schemaDef.Fields))
	for i, f := range schemaDef.Fields {
		log.Printf("[zkp]   field[%d]: name=%s public=%v", i, f.Name, f.Public)
	}

	// 3) Zbuduj assignments z VCStore (h.VCs), bez ręcznego mappingu
	assignments := make(map[string]interface{})

	vcList, err := h.VCs.List()
	if err != nil {
		log.Printf("[zkp] vcstore list failed: %v", err)
		http.Error(w, "vcstore list failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("[zkp] vcstore has %d credentials", len(vcList))
	if len(vcList) == 0 {
		http.Error(w, "no credentials in wallet – import a VC first", http.StatusBadRequest)
		return
	}

	for _, stored := range vcList {
		log.Printf("[zkp] processing VC id=%s format=%s", stored.ID, stored.Format)

		// stored.Raw = cały JWT/JWS
		msg, err := jws.Parse([]byte(stored.Raw))
		if err != nil {
			log.Printf("[zkp]   jws parse failed for VC %s: %v", stored.ID, err)
			continue
		}

		var claims map[string]any
		if err := json.Unmarshal(msg.Payload(), &claims); err != nil {
			log.Printf("[zkp]   payload json unmarshal failed for VC %s: %v", stored.ID, err)
			continue
		}

		cs := extractCredentialSubject(claims)

		for _, field := range schemaDef.Fields {
			// aud/nonce pomijamy tutaj, nadpiszemy z desc
			if field.Name == "aud" || field.Name == "nonce" {
				continue
			}
			if _, exists := assignments[field.Name]; exists {
				continue
			}

			// 1) spróbuj top-level
			if v, ok := claims[field.Name]; ok {
				assignments[field.Name] = v
				log.Printf("[zkp]   field %s resolved from top-level claims (VC %s)", field.Name, stored.ID)
				continue
			}

			// 2) spróbuj vc.credentialSubject.<name>
			if cs != nil {
				if v, ok := cs[field.Name]; ok {
					assignments[field.Name] = v
					log.Printf("[zkp]   field %s resolved from vc.credentialSubject (VC %s)", field.Name, stored.ID)
					continue
				}
			}

			// 3) spróbuj ścieżkę z kropkami, np. "vc.credentialSubject.birth_ts"
			if v, ok := getByPath(claims, field.Name); ok {
				assignments[field.Name] = v
				log.Printf("[zkp]   field %s resolved via path lookup (VC %s)", field.Name, stored.ID)
				continue
			}
		}
	}

	// 3a) aud / nonce z DI (jeśli istnieją w schemie)
	if _, err := schemaDef.FieldDefinition("aud"); err == nil && desc.Audience != "" {
		assignments["aud"] = desc.Audience
		log.Printf("[zkp]   field aud set from descriptor.Audience=%s", desc.Audience)
	}
	if _, err := schemaDef.FieldDefinition("nonce"); err == nil && desc.Nonce != "" {
		assignments["nonce"] = desc.Nonce
		log.Printf("[zkp]   field nonce set from descriptor.Nonce=%s", desc.Nonce)
	}

	// sprawdź, jak wygląda assignments (log json)
	if b, err := json.Marshal(assignments); err == nil {
		log.Printf("[zkp] assignments built: %s", string(b))
	} else {
		log.Printf("[zkp] assignments built but json.Marshal failed: %v", err)
	}

	// 3b) walidacja: czy mamy wszystkie pola z schema (poza aud/nonce)
	missing := make([]string, 0)
	for _, f := range schemaDef.Fields {
		if f.Name == "aud" || f.Name == "nonce" {
			continue
		}
		if _, ok := assignments[f.Name]; !ok {
			missing = append(missing, f.Name)
		}
	}

	if len(missing) > 0 {
		log.Printf("[zkp] no credentials available to satisfy schema %s, missing fields: %v",
			desc.Schema.Hash, missing)
		http.Error(
			w,
			"no credentials available to satisfy schema; missing fields: "+strings.Join(missing, ", "),
			http.StatusBadRequest,
		)
		return
	}

	// 4) generacja ZKP
	zkpResult, err := zkp.ProveDynamicFromSchema(schemaJSON, assignments, pkBytes)
	if err != nil {
		log.Printf("[zkp] zkp prove failed: %v", err)
		http.Error(w, "zkp prove failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5) Borsh + base64
	borshBytes, err := zkpResult.SerializeBorsh()
	if err != nil {
		log.Printf("[zkp] zkp serialize failed: %v", err)
		http.Error(w, "zkp serialize failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	b64 := base64.StdEncoding.EncodeToString(borshBytes)
	log.Printf("[zkp] proof generated, borsh_len=%d", len(borshBytes))

	// 6) public_inputs z public fields
	publicInputs := make(map[string]any)
	for _, field := range schemaDef.Fields {
		if !field.Public {
			continue
		}
		if v, ok := assignments[field.Name]; ok {
			publicInputs[field.Name] = v
		}
	}
	if b, err := json.Marshal(publicInputs); err == nil {
		log.Printf("[zkp] public_inputs: %s", string(b))
	}

	// 7) Call DI verify endpoint
	payload := map[string]any{
		"request_id":    desc.RequestID,
		"zkp_blob_b64":  b64,
		"public_inputs": publicInputs,
		"challenge":     desc.Challenge,
	}

	log.Printf("[zkp] calling verifier: %s", desc.SubmitURL)

	var verifierResp any
	if err := util.HttpPostJSON(desc.SubmitURL, payload, &verifierResp); err != nil {
		log.Printf("[zkp] digital identity verify failed: %v", err)
		http.Error(w, "digital identity verify failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("[zkp] verifier call OK for request_id=%s", desc.RequestID)

	// 8) Response do frontu
	resp := ZkpCreateResponse{
		RequestID:        desc.RequestID,
		SchemaHash:       desc.Schema.Hash,
		SchemaURI:        desc.Schema.URI,
		Audience:         desc.Audience,
		ExpiresAt:        desc.ExpiresAt,
		SubmitURL:        desc.SubmitURL,
		ZkpBorshBase64:   b64,
		ProofLength:      len(borshBytes),
		PublicWitnessOK:  true,
		VerifierResponse: verifierResp,
	}

	util.WriteJSON(w, resp)
}

// ----------------------------
//   Claims z VCStore (pretty)
// ----------------------------

func (h *ZkpHandler) HandleWalletClaims(w http.ResponseWriter, r *http.Request) {
	id := path.Base(r.URL.Path)

	vc, ok := h.VCs.Get(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	msg, err := jws.Parse([]byte(vc.Raw))
	if err != nil {
		http.Error(w, "jws parse error: "+err.Error(), http.StatusBadRequest)
		return
	}

	var claims map[string]any
	if err := json.Unmarshal(msg.Payload(), &claims); err != nil {
		http.Error(w, "payload json error: "+err.Error(), http.StatusBadRequest)
		return
	}

	util.WriteJSON(w, map[string]any{
		"id":           vc.ID,
		"format":       vc.Format,
		"display_name": vc.ID,
		"issuer_meta":  vc.Issuer,
		"subject_meta": vc.Subject,
		"types_meta":   vc.Types,
		"claims":       claims,
	})
}
