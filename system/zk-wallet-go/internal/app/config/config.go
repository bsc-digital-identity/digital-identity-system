package config

import (
	"crypto/ed25519"
	"log"
	"os"
	"sync"
	"zk-wallet-go/pkg/util/timeutil"

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
	IssuerBaseURL = "http://localhost:8080"

	// URL of the external OIDC provider (e.g., DSNet / AGH SSO)
	DsnetIssuer string

	// OAuth2 client credentials registered with the OIDC provider
	OidcClientID     string
	OidcClientSecret string // may be empty for public clients (no secret)

	// Callback URL registered with the OIDC provider
	// Typically: IssuerBaseURL + "/auth/dsnet/callback"
	RedirectURI string

	// OIDC & OAuth2 configuration (populated in setupOIDC)
	Oauth2Config *oauth2.Config
	OidcProvider *oidc.Provider
	Verifier     *oidc.IDTokenVerifier // verifies ID tokens from provider

	// Keys used by this local issuer (for signing/verifying JWT VCs)
	IssuerJWKSet  jwk.Set
	IssuerPrivKey ed25519.PrivateKey
	IssuerKeyID   string

	// Cookie names used across authentication flow
	CookieNameSession = "sid"
	CookieNamePKCE    = "pkce_verifier"
)

// ---------------------------------------------------------------
//                   SESSION / STORAGE STRUCTS
// ---------------------------------------------------------------

// Session represents a simple authenticated user Session stored in-memory.
// Fields correspond to standard OIDC claims.
type Session struct {
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
	SessStore = map[string]Session{} // active user sessions
	SessMu    sync.RWMutex

	PreAuthStore = map[string]Preauth{} // pre-authorized offers (OIDC4VCI pre-auth codes)
	PreAuthMu    sync.RWMutex

	AccessStore = map[string]AccessRec{} // issued access tokens (OIDC4VCI)
	AccessMu    sync.RWMutex

	WalletVCs = map[string]StoredVC{} // verifiable credentials "owned" by wallets
	WalletMu  sync.RWMutex
)

// ---------------------------------------------------------------
//                   DATA MODELS
// ---------------------------------------------------------------

// Preauth holds temporary pre-authorization data for OIDC4VCI
type Preauth struct {
	Code      string           // pre-authorization code
	User      Session          // user associated with it
	ExpiresAt timeutil.TimeUTC // expiration time (in UTC)
}

// AccessRec represents an access token grant record (for OIDC4VCI token flow)
type AccessRec struct {
	Token  string           // access token value
	User   Session          // user associated with the token
	CNonce string           // cryptographic nonce for proof-of-possession
	Exp    timeutil.TimeUTC // expiration timestamp
}

// StoredVC represents a verifiable credential stored in a user's wallet.
// It is identified by the JWS hash (ID) and may include metadata for display.
type StoredVC struct {
	ID          string           `json:"id"`                     // hash of the compact JWS
	Format      string           `json:"format"`                 // VC format, e.g., "jwt_vc_json"
	Credential  string           `json:"credential"`             // compact JWS representation
	Issuer      string           `json:"issuer"`                 // DID or URL of the issuer
	Subject     string           `json:"subject"`                // subject (holder) of the credential
	Types       []string         `json:"types"`                  // credential type(s)
	ReceivedAt  timeutil.TimeUTC `json:"received_at"`            // when it was issued/received
	DisplayName string           `json:"display_name,omitempty"` // optional human-readable name
}

// ---------------------------------------------------------------
//                   TIME WRAPPER
// ---------------------------------------------------------------

// ---------------------------------------------------------------
//                   ENVIRONMENT HELPERS
// ---------------------------------------------------------------

// MustEnv returns the value of an environment variable or logs a fatal error
// if it is not defined. Used for required config values (e.g., CLIENT_ID).
func MustEnv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("missing env %s", k)
	}
	return v
}

// GetenvDefault returns the environment variable value if set,
// or a provided default if not. Used for optional configuration values.
func GetenvDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
