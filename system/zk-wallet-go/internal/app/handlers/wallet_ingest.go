package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"zk-wallet-go/internal/app/config"
	"zk-wallet-go/internal/app/vcstore"
	"zk-wallet-go/pkg/util"

	"github.com/lestrrat-go/jwx/v2/jws"
)

// IngestRequest represents the input for wallet ingestion.
// Exactly one of Credential (compact JWS) or Offer (credential offer URI/deeplink) should be provided.
// The Store flag controls whether the VC should be persisted locally (defaults to true).
type IngestRequest struct {
	// Provide exactly one:
	Credential string `json:"credential,omitempty"` // compact JWS (jwt_vc_json)
	Offer      string `json:"offer,omitempty"`      // credential_offer_uri OR openid-credential-offer://...
	// Optional: whether to store (default true)
	Store *bool `json:"store,omitempty"`
}

// IngestResponse is the unified output describing what was ingested,
// whether the credential verified, basic extracted metadata, and the raw payload (best-effort).
type IngestResponse struct {
	Source  string   `json:"source"`          // "credential" | "offer"
	Valid   bool     `json:"valid"`           // signature verification result
	VCID    string   `json:"vc_id,omitempty"` // sha256 of compact JWS
	Issuer  string   `json:"issuer,omitempty"`
	Types   []string `json:"types,omitempty"`
	Stored  bool     `json:"stored"`
	Payload any      `json:"payload,omitempty"` // decoded JWT payload (best-effort)
}

// HandleWalletIngest accepts either a direct compact JWS credential or a credential offer,
// optionally redeems the offer (token → credential), verifies the VC (best-effort),
// stores it in VCStore (if requested), and returns a normalized response with metadata.
func (h *WalletHandler) HandleWalletIngest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var in IngestRequest
	// Parse incoming JSON request.
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Default to storing the VC unless explicitly disabled.
	store := true
	if in.Store != nil {
		store = *in.Store
	}

	var compact string // the resolved compact JWS to process
	var source string  // "credential" if provided directly, otherwise "offer"

	switch {
	case strings.TrimSpace(in.Credential) != "":
		// Direct credential path.
		compact = in.Credential
		source = "credential"

	case strings.TrimSpace(in.Offer) != "":
		// Offer path: normalize, fetch offer JSON, exchange code for token, request credential.
		offerURL, err := normalizeOfferToURI(in.Offer)
		if err != nil {
			http.Error(w, "invalid offer: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Get the credential offer document.
		var offer struct {
			CredentialIssuer           string                 `json:"credential_issuer"`
			CredentialConfigurationIDs []string               `json:"credential_configuration_ids"`
			Grants                     map[string]interface{} `json:"grants"`
		}
		if err := util.HttpGetJSON(offerURL, &offer); err != nil {
			http.Error(w, "offer fetch failed: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Extract the pre-authorized code grant.
		grant, _ := offer.Grants["urn:ietf:params:oauth:grant-type:pre-authorized_code"].(map[string]any)
		pre, _ := grant["pre-authorized_code"].(string)
		if pre == "" {
			http.Error(w, "offer missing pre-authorized_code", http.StatusBadRequest)
			return
		}

		// Exchange pre-authorized code for access token (MVP: without tx_code).
		var tokRes struct {
			AccessToken string `json:"access_token"`
			TokenType   string `json:"token_type"`
			ExpiresIn   int    `json:"expires_in"`
			CNonce      string `json:"c_nonce"`
		}
		if err := util.HttpPostJSON(offer.CredentialIssuer+"/oidc4vci/token", map[string]any{
			"grant_type":          "urn:ietf:params:oauth:grant-type:pre-authorized_code",
			"pre-authorized_code": pre,
		}, &tokRes); err != nil {
			http.Error(w, "token failed: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Request the actual credential (MVP: no PoP proof yet).
		var credRes struct {
			Format     string `json:"format"`
			Credential string `json:"credential"`
		}
		if err := util.HttpPostAuthJSON(offer.CredentialIssuer+"/oidc4vci/credential", tokRes.AccessToken, map[string]any{}, &credRes); err != nil {
			http.Error(w, "credential failed: "+err.Error(), http.StatusBadRequest)
			return
		}

		compact = credRes.Credential
		source = "offer"

	default:
		// Neither direct credential nor offer provided → bail out.
		http.Error(w, "provide either 'credential' or 'offer'", http.StatusBadRequest)
		return
	}

	// Verify signature against local issuerJWKSet (PoC). In production, resolve issuer JWKS.
	valid := true
	msg, err := jws.Verify([]byte(compact), jws.WithKeySet(config.IssuerJWKSet))
	if err != nil {
		valid = false // If verification fails, we still proceed with best-effort decode.
	}

	// Best-effort decode of payload for transparency/debugging.
	var payload any
	_ = json.Unmarshal(msg, &payload)

	// Extract a few common fields (issuer, subject, credential types).
	var pl struct {
		Iss string `json:"iss"`
		Sub string `json:"sub"`
		VC  struct {
			Type []string `json:"type"`
		} `json:"vc"`
	}
	_ = json.Unmarshal(msg, &pl)

	// Compute a stable VC ID as sha256(Base64URL JWS).
	vcID := util.Sha256hex([]byte(compact))
	stored := false

	// Optionally store the credential in VCStore.
	if store {
		vc := vcstore.VerifiableCredential{
			ID:        vcID,
			Format:    "jwt_vc_json",
			Raw:       compact,
			Issuer:    pl.Iss,
			Subject:   pl.Sub,
			Types:     pl.VC.Type,
			CreatedAt: time.Now().UTC(),
		}

		if err := h.VCs.Save(vc); err != nil {
			http.Error(w, "vcstore save failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		stored = true
	}

	// Return a normalized response with verification result and parsed info.
	util.WriteJSON(w, IngestResponse{
		Source:  source,
		Valid:   valid,
		VCID:    vcID,
		Issuer:  pl.Iss,
		Types:   pl.VC.Type,
		Stored:  stored,
		Payload: payload,
	})
}

// normalizeOfferToURI converts supported inputs to a standard credential_offer_uri:
// - If input is a deeplink (openid-credential-offer://?...), extract credential_offer_uri from its query.
// - If input is already an http(s) URL, validate scheme and return as-is.
func normalizeOfferToURI(input string) (string, error) {
	// Accept:
	//  - credential_offer_uri (http/https URL)
	//  - openid-credential-offer://?credential_offer_uri=...
	if strings.HasPrefix(input, "openid-credential-offer://") {
		u, err := util.ParseURL(input)
		if err != nil {
			return "", err
		}
		v := u.Query().Get("credential_offer_uri")
		if v == "" {
			return "", errBadOffer
		}
		input = v
	}

	// Validate that the remaining input is an http(s) URL.
	u, err := url.Parse(input)
	if err != nil {
		return "", err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errBadOffer
	}
	return input, nil
}

// errBadOffer is a sentinel error returned for invalid offer inputs.
var errBadOffer = &badOfferErr{}

// badOfferErr implements error for invalid credential offer inputs.
type badOfferErr struct{}

func (e *badOfferErr) Error() string { return "invalid credential offer input" }
