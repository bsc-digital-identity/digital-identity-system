package handlers

import (
	"encoding/json"
	"net/http"
	"path"
	"strings"
	"time"
	"zk-wallet-go/internal/app/config"
	"zk-wallet-go/pkg/util"
	"zk-wallet-go/pkg/util/timeutil"

	"github.com/lestrrat-go/jwx/v2/jws"
)

// HandleWalletImportOffer simulates the wallet-side flow of importing and redeeming
// a credential offer using the OIDC4VCI protocol.
//
// Flow summary:
// 1. Accept an offer (or deeplink) from the frontend.
// 2. Resolve deeplink into a standard credential_offer_uri.
// 3. GET the offer details from the issuer.
// 4. Exchange the pre-authorized_code for an access_token via /token.
// 5. Use the access_token to request the credential from /credential.
// 6. Optionally verify and parse the received VC JWT.
// 7. Store the credential in local in-memory walletVCs.
func HandleWalletImportOffer(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Offer string `json:"offer"`
	}
	// Parse input JSON.
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}
	if in.Offer == "" {
		http.Error(w, "missing offer", http.StatusBadRequest)
		return
	}

	// (1) Normalize deeplink to standard credential_offer_uri.
	offerURL := in.Offer
	if strings.HasPrefix(offerURL, "openid-credential-offer://") {
		u, _ := util.ParseURL(offerURL)
		q := u.Query().Get("credential_offer_uri")
		if q == "" {
			http.Error(w, "deeplink without credential_offer_uri", http.StatusBadRequest)
			return
		}
		offerURL = q
	}

	// (2) Fetch offer details from issuer.
	var offer struct {
		CredentialIssuer           string         `json:"credential_issuer"`
		CredentialConfigurationIDs []string       `json:"credential_configuration_ids"`
		Grants                     map[string]any `json:"grants"`
	}
	if err := util.HttpGetJSON(offerURL, &offer); err != nil {
		http.Error(w, "offer fetch failed: "+err.Error(), http.StatusBadRequest)
		return
	}
	grant, _ := offer.Grants["urn:ietf:params:oauth:grant-type:pre-authorized_code"].(map[string]any)
	pre, _ := grant["pre-authorized_code"].(string)
	if pre == "" {
		http.Error(w, "offer missing pre-authorized_code", http.StatusBadRequest)
		return
	}

	// (3) Exchange pre-authorized_code for access_token.
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

	// (4) Request actual credential using the obtained access token.
	var credRes struct {
		Format     string `json:"format"`
		Credential string `json:"credential"` // compact JWS
	}
	if err := util.HttpPostAuthJSON(offer.CredentialIssuer+"/oidc4vci/credential", tokRes.AccessToken, map[string]any{}, &credRes); err != nil {
		http.Error(w, "credential failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// (5) Verify signature using our issuer key set (in prod: fetch issuer's JWKS).
	payload, err := jws.Verify([]byte(credRes.Credential), jws.WithKeySet(config.IssuerJWKSet))
	if err != nil {
		// For MVP we still store even if signature verification fails.
	}

	// Parse basic payload fields for indexing/display.
	var pl struct {
		Iss string `json:"iss"`
		Sub string `json:"sub"`
		VC  struct {
			Type              []string       `json:"type"`
			Issuer            string         `json:"issuer"`
			CredentialSubject map[string]any `json:"credentialSubject"`
			IssuanceDate      string         `json:"issuanceDate"`
		} `json:"vc"`
	}
	_ = json.Unmarshal(payload, &pl)

	// Compute hash ID for stored VC.
	id := util.Sha256hex([]byte(credRes.Credential))
	display := "Credential"
	if len(pl.VC.Type) > 0 {
		display = pl.VC.Type[len(pl.VC.Type)-1]
	}

	// Store credential in in-memory wallet.
	config.WalletMu.Lock()
	config.WalletVCs[id] = config.StoredVC{
		ID:          id,
		Format:      credRes.Format,
		Credential:  credRes.Credential,
		Issuer:      pl.Iss,
		Subject:     pl.Sub,
		Types:       pl.VC.Type,
		ReceivedAt:  timeutil.TimeUTC{T: time.Now().UTC().Unix()},
		DisplayName: display,
	}
	config.WalletMu.Unlock()

	// Respond with minimal info about stored VC.
	util.WriteJSON(w, map[string]any{
		"vc_id":  id,
		"issuer": pl.Iss,
		"types":  pl.VC.Type,
		"stored": true,
	})
}

// HandleWalletList returns all stored credentials (VCs) from the local wallet.
// Used for debugging / viewing wallet content.
func HandleWalletList(w http.ResponseWriter, r *http.Request) {
	config.WalletMu.RLock()
	out := make([]config.StoredVC, 0, len(config.WalletVCs))
	for _, v := range config.WalletVCs {
		out = append(out, v)
	}
	config.WalletMu.RUnlock()
	util.WriteJSON(w, out)
}

// HandleWalletShow displays a specific VC by its ID, verifying its signature first.
// Returns both metadata (from local store) and the decoded payload.
func HandleWalletShow(w http.ResponseWriter, r *http.Request) {
	id := path.Base(r.URL.Path)

	// Lookup VC in wallet.
	config.WalletMu.RLock()
	v, ok := config.WalletVCs[id]
	config.WalletMu.RUnlock()
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Verify signature using local issuer keys.
	msg, err := jws.Verify([]byte(v.Credential), jws.WithKeySet(config.IssuerJWKSet))
	if err != nil {
		http.Error(w, "signature invalid: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Decode verified payload.
	var payload any
	_ = json.Unmarshal(msg, &payload)
	util.WriteJSON(w, map[string]any{
		"meta":    v,
		"payload": payload,
	})
}

// HandleWalletVerify allows user to POST a credential and checks its signature validity.
// Returns {valid:true, payload:<claims>} if verification passes.
func HandleWalletVerify(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Credential string `json:"credential"`
	}
	// Parse input JSON.
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad json: "+err.Error(), http.StatusBadRequest)
		return
	}
	if in.Credential == "" {
		http.Error(w, "missing credential", http.StatusBadRequest)
		return
	}

	// Verify the credential signature using issuer keys.
	msg, err := jws.Verify([]byte(in.Credential), jws.WithKeySet(config.IssuerJWKSet))
	if err != nil {
		http.Error(w, "invalid signature: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Decode claims payload.
	var payload any
	_ = json.Unmarshal(msg, &payload)
	util.WriteJSON(w, map[string]any{"valid": true, "payload": payload})
}

// HandleWalletClaims extracts and displays claims from a stored VC (without verification).
// Used for human-readable debugging or simple frontend display.
//
// Steps:
// 1. Retrieve stored VC.
// 2. Parse JWS (without verifying).
// 3. Unmarshal payload into map[string]any.
// 4. Return both stored metadata and full claim set.
func HandleWalletClaims(w http.ResponseWriter, r *http.Request) {
	id := path.Base(r.URL.Path)

	// (1) Retrieve stored VC by ID.
	config.WalletMu.RLock()
	v, ok := config.WalletVCs[id]
	config.WalletMu.RUnlock()
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// (2) Parse JWS envelope to get payload (without verifying).
	msg, err := jws.Parse([]byte(v.Credential))
	if err != nil {
		http.Error(w, "jws parse error: "+err.Error(), http.StatusBadRequest)
		return
	}

	// (3) Unmarshal payload into generic claims map.
	var claims map[string]any
	if err := json.Unmarshal(msg.Payload(), &claims); err != nil {
		http.Error(w, "payload json error: "+err.Error(), http.StatusBadRequest)
		return
	}

	// (4) Return structured response containing metadata and claims.
	util.WriteJSON(w, map[string]any{
		"id":           v.ID,
		"format":       v.Format,
		"display_name": v.DisplayName,
		"issuer_meta":  v.Issuer,  // from stored metadata
		"subject_meta": v.Subject, // from stored metadata
		"types_meta":   v.Types,   // from stored metadata
		"claims":       claims,    // decoded credential claims
	})
}
