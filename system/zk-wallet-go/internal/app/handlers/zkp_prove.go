// zkp_prove.go
package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"zk-wallet-go/internal/app/config"
	"zk-wallet-go/internal/app/zkp"
	"zk-wallet-go/pkg/util"
)

// fetchDescriptorAndSchema pobiera descriptor + schema JSON dla danego request_id.
func fetchDescriptorSchemaAndPK(requestID string) (PresentationDescriptor, []byte, []byte, error) {
	var desc PresentationDescriptor

	base := strings.TrimRight(config.ZkpVerifierBaseURL, "/")
	descURL := base + "/v1/presentations/" + requestID + "/descriptor"

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

	desc.Schema.URI = normalizeWeirdURL(desc.Schema.URI)
	desc.SubmitURL = normalizeWeirdURL(desc.SubmitURL)
	desc.Artifacts.VKURL = normalizeWeirdURL(desc.Artifacts.VKURL)
	desc.Artifacts.PKURL = normalizeWeirdURL(desc.Artifacts.PKURL)
	desc.Audience = normalizeWeirdURL(desc.Audience)

	if strings.TrimSpace(desc.Schema.URI) == "" {
		return desc, nil, nil, fmt.Errorf("descriptor has empty schema.uri")
	}
	if strings.TrimSpace(desc.Artifacts.PKURL) == "" {
		return desc, nil, nil, fmt.Errorf("descriptor has empty artifacts.pk_url")
	}

	// fetch schema
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

	return desc, schemaBytes, pkBytes, nil
}

// ----------------------------
//   API: /wallet/zkp/create
// ----------------------------

type ZkpCreateRequest struct {
	// request_id z DI (np. z QR / linku)
	RequestID string `json:"request_id"`
	// mapowanie: nazwa_pola_z_schema -> wartość użytkownika
	Inputs map[string]interface{} `json:"inputs,omitempty"`
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

func HandleZkpCreate(w http.ResponseWriter, r *http.Request) {
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

	// 1) descriptor + schema + pk
	desc, schemaJSON, pkBytes, err := fetchDescriptorSchemaAndPK(reqID)
	if err != nil {
		http.Error(w, "descriptor/schema/pk fetch failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 2) assignments z inputs
	assignments := map[string]interface{}{}
	for k, v := range in.Inputs {
		assignments[k] = v
	}

	// 3) parse schema
	schemaDef, err := zkp.ParseSchema(schemaJSON)
	if err != nil {
		http.Error(w, "schema parse failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// 3a) aud / nonce z DI (jeśli istnieją w schemie)
	if _, err := schemaDef.FieldDefinition("aud"); err == nil && desc.Audience != "" {
		if _, ok := assignments["aud"]; !ok {
			assignments["aud"] = desc.Audience
		}
	}
	if _, err := schemaDef.FieldDefinition("nonce"); err == nil && desc.Nonce != "" {
		if _, ok := assignments["nonce"]; !ok {
			assignments["nonce"] = desc.Nonce
		}
	}

	// 4) generacja ZKP – teraz z pkBytes z DI
	zkpResult, err := zkp.ProveDynamicFromSchema(schemaJSON, assignments, pkBytes)
	if err != nil {
		http.Error(w, "zkp prove failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 5) Borsh + base64
	borshBytes, err := zkpResult.SerializeBorsh()
	if err != nil {
		http.Error(w, "zkp serialize failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	b64 := base64.StdEncoding.EncodeToString(borshBytes)

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

	// 7) Call DI verify endpoint
	payload := map[string]any{
		"request_id":    desc.RequestID,
		"zkp_blob_b64":  b64,
		"public_inputs": publicInputs,
		"challenge":     desc.Challenge,
	}

	var verifierResp any
	if err := util.HttpPostJSON(desc.SubmitURL, payload, &verifierResp); err != nil {
		http.Error(w, "digital identity verify failed: "+err.Error(), http.StatusBadRequest)
		return
	}

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
