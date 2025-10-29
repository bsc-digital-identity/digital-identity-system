package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"path"
	"strings"
	"time"
	"zk-wallet-go/internal/app/config"
	"zk-wallet-go/internal/app/server"
	"zk-wallet-go/pkg/util"
	"zk-wallet-go/pkg/util/timeutil"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jws"
)

// HandleCreateOffer creates a new OIDC4VCI "credential offer".
// This endpoint is typically called by an authenticated user (issuer UI) to generate
// a short-lived offer code and a deeplink for wallet apps.
//
// Flow summary:
// 1. Verify user session.
// 2. Generate random pre-authorized code valid for 5 minutes.
// 3. Return both credential_offer_uri and deeplink for wallet scanning.
func HandleCreateOffer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Require active login session.
	s, err := server.CurrentSession(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Create a short-lived pre-authorization record (5 min lifetime).
	code := util.RandomString(32)
	config.PreAuthMu.Lock()
	config.PreAuthStore[code] = config.Preauth{Code: code, User: s, ExpiresAt: timeutil.NowUTC().AddSeconds(300)}
	config.PreAuthMu.Unlock()

	// Construct URLs for wallet consumption.
	offerURI := config.IssuerBaseURL + "/oidc4vci/offer/" + code
	deeplink := "openid-credential-offer://?credential_offer_uri=" + urlQueryEscape(offerURI)

	// Respond with both the API URL and the mobile deeplink.
	resp := map[string]any{
		"credential_offer_uri": offerURI,
		"deeplink":             deeplink,
	}
	util.WriteJSON(w, resp)
}

// HandleOfferByCode serves the actual credential offer object for wallets
// when they resolve the credential_offer_uri (GET /oidc4vci/offer/{code}).
// It validates that the offer code exists and has not expired, then returns
// the standard credential_offer JSON payload per OIDC4VCI spec.
func HandleOfferByCode(w http.ResponseWriter, r *http.Request) {
	code := path.Base(r.URL.Path)

	// Lookup offer and validate expiration.
	config.PreAuthMu.RLock()
	pa, ok := config.PreAuthStore[code]
	config.PreAuthMu.RUnlock()
	if !ok || timeutil.NowUTC().After(pa.ExpiresAt) {
		http.Error(w, "offer expired or invalid", http.StatusBadRequest)
		return
	}

	// Respond with credential offer payload.
	offer := map[string]any{
		"credential_issuer":            config.IssuerBaseURL,
		"credential_configuration_ids": []string{"StudentCredential_JWT_v1"},
		"grants": map[string]any{
			"urn:ietf:params:oauth:grant-type:pre-authorized_code": map[string]any{
				"pre-authorized_code": code,
			},
		},
	}
	util.WriteJSON(w, offer)
}

// HandleVciToken implements the OIDC4VCI /token endpoint.
// The wallet calls this endpoint with the pre-authorized code to exchange it
// for an access_token and c_nonce used in credential issuance.
//
// Flow summary:
// 1. Parse JSON body and check grant_type.
// 2. Validate pre-authorized code and expiration.
// 3. Issue short-lived access_token (5 min) and c_nonce.
// 4. Return token response per OIDC4VCI spec.
func HandleVciToken(w http.ResponseWriter, r *http.Request) {
	log.Println("[VCI] /token hit")

	// Decode request body.
	var body struct {
		GrantType         string `json:"grant_type"`
		PreAuthorizedCode string `json:"pre-authorized_code"`
		TxCode            string `json:"tx_code,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Println("[VCI] bad json:", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	log.Printf("[VCI] grant=%s precode=%s\n", body.GrantType, body.PreAuthorizedCode)
	// Validate supported grant type.
	if body.GrantType != "urn:ietf:params:oauth:grant-type:pre-authorized_code" {
		http.Error(w, "unsupported grant", http.StatusBadRequest)
		return
	}

	// Validate pre-authorized code and its expiration.
	config.PreAuthMu.RLock()
	pa, ok := config.PreAuthStore[body.PreAuthorizedCode]
	config.PreAuthMu.RUnlock()
	if !ok || timeutil.NowUTC().After(pa.ExpiresAt) {
		http.Error(w, "invalid or expired pre-authorized code", http.StatusBadRequest)
		return
	}

	// Create a new short-lived access token record (5 min).
	at := "atk_" + util.RandomString(32)
	cnonce := util.RandomString(32)
	config.AccessMu.Lock()
	config.AccessStore[at] = config.AccessRec{
		Token:  at,
		User:   pa.User,
		CNonce: cnonce,
		Exp:    timeutil.NowUTC().AddSeconds(300),
	}
	config.AccessMu.Unlock()

	// Invalidate (single-use) pre-authorized code.
	config.PreAuthMu.Lock()
	delete(config.PreAuthStore, body.PreAuthorizedCode)
	config.PreAuthMu.Unlock()

	// Return token response.
	resp := map[string]any{
		"access_token":       at,
		"token_type":         "bearer",
		"expires_in":         300,
		"c_nonce":            cnonce,
		"c_nonce_expires_in": 300,
	}
	util.WriteJSON(w, resp)
}

// HandleVciCredential issues a verifiable credential as a signed JWT (VC-JWT format).
// The wallet calls this endpoint with the access_token (and normally a proof JWT).
//
// Flow summary:
// 1. Validate bearer token and expiration.
// 2. (Normally) verify wallet proof-of-possession using c_nonce.
// 3. Create a VC payload with example claims.
// 4. Sign using issuer's Ed25519 private key.
// 5. Return the signed credential in VC-JWT format.
func HandleVciCredential(w http.ResponseWriter, r *http.Request) {
	// Check Authorization header.
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	at := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))

	// Validate token existence and expiration.
	config.AccessMu.RLock()
	rec, ok := config.AccessStore[at]
	config.AccessMu.RUnlock()
	if !ok || timeutil.NowUTC().After(rec.Exp) {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	// NOTE: In production, verify PoP JWT from wallet contains rec.CNonce and is signed by holder key.

	now := time.Now().UTC()

	// Example static claims — in a real system these would come from DSNet or your identity source.
	studentStatus := "active"
	studentID := "123456"
	gender := "unspecified"
	birthdate := "2004-05-10"

	// Compose VC-JWT claims.
	vcClaims := map[string]any{
		"iss": config.IssuerBaseURL,
		"sub": rec.User.Sub,
		"iat": now.Unix(),
		"nbf": now.Unix(),
		"vc": map[string]any{
			"@context":     []any{"https://www.w3.org/2018/credentials/v1"},
			"type":         []string{"VerifiableCredential", "StudentCredential"},
			"issuer":       config.IssuerBaseURL,
			"issuanceDate": now.Format(time.RFC3339),
			"credentialSubject": map[string]any{
				"id":             rec.User.Sub,
				"student_status": studentStatus,
				"student_id":     studentID,
				"gender":         gender,
				"birthdate":      birthdate,
			},
		},
	}

	// Marshal claims to JSON for signing.
	payload, _ := json.Marshal(vcClaims)

	// Prepare protected headers for JWS.
	hdr := jws.NewHeaders()
	_ = hdr.Set(jws.AlgorithmKey, jwa.EdDSA)
	_ = hdr.Set(jws.KeyIDKey, config.IssuerKeyID)
	_ = hdr.Set("typ", "vc+jwt")

	// Sign payload with issuer's Ed25519 private key.
	signed, err := jws.Sign(
		payload,
		jws.WithKey(
			jwa.EdDSA,
			config.IssuerPrivKey,
			jws.WithProtectedHeaders(hdr),
		),
	)
	if err != nil {
		http.Error(w, "sign failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return credential in OIDC4VCI standard response format.
	resp := map[string]any{
		"format":     "jwt_vc_json",
		"credential": string(signed),
	}
	util.WriteJSON(w, resp)
}

// tiny helpers

// urlQueryEscape wraps escapeQuery (local implementation elsewhere) to URL-encode query parameters.
func urlQueryEscape(s string) string { return util.EscapeQuery(s) }
