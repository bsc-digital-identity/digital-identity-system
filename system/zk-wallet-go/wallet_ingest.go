package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jws"
)

// ingestRequest represents the input for wallet ingestion.
// Exactly one of Credential (compact JWS) or Offer (credential offer URI/deeplink) should be provided.
// The Store flag controls whether the VC should be persisted locally (defaults to true).
type ingestRequest struct {
	// Provide exactly one:
	Credential string `json:"credential,omitempty"` // compact JWS (jwt_vc_json)
	Offer      string `json:"offer,omitempty"`      // credential_offer_uri OR openid-credential-offer://...
	// Optional: whether to store (default true)
	Store *bool `json:"store,omitempty"`
}

// ingestResponse is the unified output describing what was ingested,
// whether the credential verified, basic extracted metadata, and the raw payload (best-effort).
type ingestResponse struct {
	Source  string   `json:"source"`          // "credential" | "offer"
	Valid   bool     `json:"valid"`           // signature verification result
	VCID    string   `json:"vc_id,omitempty"` // sha256 of compact JWS
	Issuer  string   `json:"issuer,omitempty"`
	Types   []string `json:"types,omitempty"`
	Stored  bool     `json:"stored"`
	Payload any      `json:"payload,omitempty"` // decoded JWT payload (best-effort)
}

// handleWalletIngest accepts either a direct compact JWS credential or a credential offer,
// optionally redeems the offer (token → credential), verifies the VC (best-effort),
// stores it (if requested), and returns a normalized response with metadata.
func handleWalletIngest(w http.ResponseWriter, r *http.Request) {
	var in ingestRequest
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
		if err := httpGetJSON(offerURL, &offer); err != nil {
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
		if err := httpPostJSON(offer.CredentialIssuer+"/oidc4vci/token", map[string]any{
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
		if err := httpPostAuthJSON(offer.CredentialIssuer+"/oidc4vci/credential", tokRes.AccessToken, map[string]any{}, &credRes); err != nil {
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
	msg, err := jws.Verify([]byte(compact), jws.WithKeySet(issuerJWKSet))
	if err != nil {
		valid = false // If verification fails, we still proceed with best-effort decode.
		// best-effort decode will likely be empty if verify failed
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
	vcID := sha256hex([]byte(compact))
	stored := false

	// Optionally store the credential in the in-memory wallet cache.
	if store {
		// Choose a display name (last declared type if present).
		display := "Credential"
		if len(pl.VC.Type) > 0 {
			display = pl.VC.Type[len(pl.VC.Type)-1]
		}

		walletMu.Lock()
		// Avoid overwriting if the VC already exists.
		if _, exists := walletVCs[vcID]; !exists {
			walletVCs[vcID] = StoredVC{
				ID:          vcID,
				Format:      "jwt_vc_json",
				Credential:  compact,
				Issuer:      pl.Iss,
				Subject:     pl.Sub,
				Types:       pl.VC.Type,
				ReceivedAt:  TimeUTC{T: time.Now().UTC().Unix()},
				DisplayName: display,
			}
		}
		stored = true
		walletMu.Unlock()
	}

	// Return a normalized response with verification result and parsed info.
	writeJSON(w, ingestResponse{
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
		u, err := parseURL(input)
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
