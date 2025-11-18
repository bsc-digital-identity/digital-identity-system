package handlers

import (
	"encoding/json"
	"net/http"
	"path"
	"strings"
	"time"

	"zk-wallet-go/internal/app/config"
	"zk-wallet-go/internal/app/vcstore"
	"zk-wallet-go/pkg/util"

	"github.com/lestrrat-go/jwx/v2/jws"
)

// WalletHandler – wszystkie endpointy portfela oparte o VCStore.
type WalletHandler struct {
	VCs vcstore.VCStore
}

func NewWalletHandler(store vcstore.VCStore) *WalletHandler {
	return &WalletHandler{VCs: store}
}

// HandleWalletImportOffer simuluje stronę walleta w OIDC4VCI.
//
// Flow:
// 1. Przyjmuje offer / deeplink.
// 2. Rozwiązuje deeplink do credential_offer_uri.
// 3. GET offer z issuer.
// 4. Wymienia pre-authorized_code na access_token.
// 5. Pobiera credential z /credential.
// 6. Opcjonalnie weryfikuje podpis.
// 7. Zapisuje credential do VCStore.
func (h *WalletHandler) HandleWalletImportOffer(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Offer string `json:"offer"`
	}

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
	if err := util.HttpPostAuthJSON(
		offer.CredentialIssuer+"/oidc4vci/credential",
		tokRes.AccessToken,
		map[string]any{},
		&credRes,
	); err != nil {
		http.Error(w, "credential failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	// (5) Verify signature using our issuer key set (in prod: fetch issuer's JWKS).
	payload, err := jws.Verify([]byte(credRes.Credential), jws.WithKeySet(config.IssuerJWKSet))
	if err != nil {
		// For MVP we still store even if signature verification fails.
		// payload może być nil => parsowanie niżej po prostu da zero-values.
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

	// Store credential in VCStore.
	vc := vcstore.VerifiableCredential{
		ID:      id,
		Format:  credRes.Format,
		Raw:     credRes.Credential,
		Issuer:  pl.Iss,
		Subject: pl.Sub,
		Types:   pl.VC.Type,
		// możesz tu dać time.Now().UTC() albo time.Now()
		CreatedAt: time.Now().UTC(),
	}

	if err := h.VCs.Save(vc); err != nil {
		http.Error(w, "vcstore save failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with minimal info about stored VC.
	util.WriteJSON(w, map[string]any{
		"vc_id":  id,
		"issuer": pl.Iss,
		"types":  pl.VC.Type,
		"stored": true,
	})
}

// HandleWalletList returns all stored credentials (VCs) from the local wallet.
// Uses VCStore.List().
func (h *WalletHandler) HandleWalletList(w http.ResponseWriter, r *http.Request) {
	list, err := h.VCs.List()
	if err != nil {
		http.Error(w, "vcstore list failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	util.WriteJSON(w, list)
}

// HandleWalletShow displays a specific VC by its ID, verifying its signature first.
// Returns both metadata (from VCStore) and the decoded payload.
func (h *WalletHandler) HandleWalletShow(w http.ResponseWriter, r *http.Request) {
	id := path.Base(r.URL.Path)

	vc, ok := h.VCs.Get(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Verify signature using local issuer keys.
	msg, err := jws.Verify([]byte(vc.Raw), jws.WithKeySet(config.IssuerJWKSet))
	if err != nil {
		http.Error(w, "signature invalid: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Decode verified payload.
	var payload any
	_ = json.Unmarshal(msg, &payload)

	util.WriteJSON(w, map[string]any{
		"meta":    vc,
		"payload": payload,
	})
}

// HandleWalletVerify allows user to POST a credential and checks its signature validity.
// Returns {valid:true, payload:<claims>} if verification passes.
func (h *WalletHandler) HandleWalletVerify(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Credential string `json:"credential"`
	}

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

	util.WriteJSON(w, map[string]any{
		"valid":   true,
		"payload": payload,
	})
}
