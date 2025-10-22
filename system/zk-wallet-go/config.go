package main

import (
	"crypto/ed25519"
	"log"
	"os"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"golang.org/x/oauth2"
)

// ---------------------------------------------------------------
//                   GLOBAL CONFIGURATION
// ---------------------------------------------------------------

// These global variables configure OIDC, JWT issuer identity, and in-memory stores
// for sessions and wallet data. For simplicity, this is an in-memory proof-of-concept
// (PoC) — in production, these values would be stored securely and persisted in a DB.
var (
	// Public origin of this issuer server (used for metadata, redirect URI, etc.)
	issuerBaseURL = "http://localhost:8080"

	// URL of the external OIDC provider (e.g., DSNet / AGH SSO)
	dsnetIssuer string

	// OAuth2 client credentials registered with the OIDC provider
	oidcClientID     string
	oidcClientSecret string // may be empty for public clients (no secret)

	// Callback URL registered with the OIDC provider
	// Typically: issuerBaseURL + "/auth/dsnet/callback"
	redirectURI string

	// OIDC & OAuth2 configuration (populated in setupOIDC)
	oauth2Config *oauth2.Config
	oidcProvider *oidc.Provider
	verifier     *oidc.IDTokenVerifier // verifies ID tokens from provider

	// Keys used by this local issuer (for signing/verifying JWT VCs)
	issuerJWKSet  jwk.Set
	issuerPrivKey ed25519.PrivateKey
	issuerKeyID   string

	// Cookie names used across authentication flow
	cookieNameSession = "sid"
	cookieNamePKCE    = "pkce_verifier"
)

// ---------------------------------------------------------------
//                   SESSION / STORAGE STRUCTS
// ---------------------------------------------------------------

// session represents a simple authenticated user session stored in-memory.
// Fields correspond to standard OIDC claims.
type session struct {
	Sub   string `json:"sub"`             // Subject (unique user ID)
	Email string `json:"email,omitempty"` // Optional user email
	Name  string `json:"name,omitempty"`  // Optional display name
}

// ---------------------------------------------------------------
//                   IN-MEMORY DATA STORES
// ---------------------------------------------------------------

// For this MVP, we use in-memory maps protected by RWMutexes.
// In production, these would be replaced by persistent database tables.
var (
	sessStore = map[string]session{} // active user sessions
	sessMu    sync.RWMutex

	preAuthStore = map[string]preauth{} // pre-authorized offers (OIDC4VCI pre-auth codes)
	preAuthMu    sync.RWMutex

	accessStore = map[string]accessRec{} // issued access tokens (OIDC4VCI)
	accessMu    sync.RWMutex

	walletVCs = map[string]StoredVC{} // verifiable credentials "owned" by wallets
	walletMu  sync.RWMutex
)

// ---------------------------------------------------------------
//                   DATA MODELS
// ---------------------------------------------------------------

// preauth holds temporary pre-authorization data for OIDC4VCI
type preauth struct {
	Code      string  // pre-authorization code
	User      session // user associated with it
	ExpiresAt TimeUTC // expiration time (in UTC)
}

// accessRec represents an access token grant record (for OIDC4VCI token flow)
type accessRec struct {
	Token  string  // access token value
	User   session // user associated with the token
	CNonce string  // cryptographic nonce for proof-of-possession
	Exp    TimeUTC // expiration timestamp
}

// StoredVC represents a verifiable credential stored in a user's wallet.
// It is identified by the JWS hash (ID) and may include metadata for display.
type StoredVC struct {
	ID          string   `json:"id"`                     // hash of the compact JWS
	Format      string   `json:"format"`                 // VC format, e.g., "jwt_vc_json"
	Credential  string   `json:"credential"`             // compact JWS representation
	Issuer      string   `json:"issuer"`                 // DID or URL of the issuer
	Subject     string   `json:"subject"`                // subject (holder) of the credential
	Types       []string `json:"types"`                  // credential type(s)
	ReceivedAt  TimeUTC  `json:"received_at"`            // when it was issued/received
	DisplayName string   `json:"display_name,omitempty"` // optional human-readable name
}

// ---------------------------------------------------------------
//                   TIME WRAPPER
// ---------------------------------------------------------------

// TimeUTC is a small helper type representing Unix time (in seconds) in UTC.
// Using a dedicated type prevents confusion between local and UTC timestamps.
type TimeUTC struct{ T int64 }

// ---------------------------------------------------------------
//                   ENVIRONMENT HELPERS
// ---------------------------------------------------------------

// mustEnv returns the value of an environment variable or logs a fatal error
// if it is not defined. Used for required config values (e.g., CLIENT_ID).
func mustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("missing env %s", k)
	}
	return v
}

// getenvDefault returns the environment variable value if set,
// or a provided default if not. Used for optional configuration values.
func getenvDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
