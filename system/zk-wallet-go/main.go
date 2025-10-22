package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func main() {
	_ = godotenv.Load()

	issuerBaseURL = getenvDefault("ISSUER_BASE_URL", issuerBaseURL)
	dsnetIssuer = mustEnv("DSNET_ISSUER")
	oidcClientID = mustEnv("OIDC_CLIENT_ID")
	oidcClientSecret = os.Getenv("OIDC_CLIENT_SECRET")
	redirectURI = getenvDefault("OIDC_REDIRECT_URI", issuerBaseURL+"/auth/dsnet/callback")

	// OIDC discovery
	var err error
	oidcProvider, err = oidc.NewProvider(context.Background(), dsnetIssuer)
	if err != nil {
		log.Fatalf("OIDC discover failed: %v", err)
	}
	verifier = oidcProvider.Verifier(&oidc.Config{ClientID: oidcClientID})
	oauth2Config = &oauth2.Config{
		ClientID:     oidcClientID,
		ClientSecret: oidcClientSecret,
		Endpoint:     oidcProvider.Endpoint(),
		Scopes:       []string{"openid", "email", "profile"},
		RedirectURL:  redirectURI,
	}

	log.Printf("[CONFIG] DSNET issuer: %s", dsnetIssuer)
	log.Printf("[CONFIG] Redirect URI: %s", redirectURI)
	log.Printf("[CONFIG] Client ID: %s", oidcClientID)
	log.Printf("[CONFIG] Using secret? %v", oidcClientSecret != "")
	log.Printf("[CONFIG] Token endpoint: %s", oauth2Config.Endpoint.TokenURL)

	// Issuer keypair + JWKS
	generateIssuerKey()

	// Router
	mux := http.NewServeMux()

	// Static
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/", fs)

	// --- DSNET login (OIDC code flow with PKCE) ---
	mux.HandleFunc("/auth/dsnet/login", handleLogin)       // starts the flow (redirects to DSNet)
	mux.HandleFunc("/auth/dsnet/callback", handleCallback) // handles code exchange + session creation
	mux.HandleFunc("/auth/logout", handleLogout)           // clears session

	// --- OIDC4VCI issuer metadata & JWKS ---
	mux.HandleFunc("/.well-known/openid-credential-issuer", handleCredentialIssuerMetadata) // issuer metadata
	mux.HandleFunc("/.well-known/jwks.json", handleJWKS)                                    // public keys

	// --- OIDC4VCI protocol endpoints ---
	mux.HandleFunc("/oidc4vci/offer", handleCreateOffer)        // POST: create pre-auth offer → returns deeplink/URL
	mux.HandleFunc("/oidc4vci/offer/", handleOfferByCode)       // GET:  retrieve offer by code (wallet UX)
	mux.HandleFunc("/oidc4vci/token", handleVciToken)           // POST: token exchange for credential issuance
	mux.HandleFunc("/oidc4vci/credential", handleVciCredential) // POST: issue credential (PoP-verified)

	// --- Wallet MVP endpoints ---
	mux.HandleFunc("/wallet/import-offer", handleWalletImportOffer) // POST: import an offer into wallet
	mux.HandleFunc("/wallet/vcs", handleWalletList)                 // GET:  list stored VCs
	mux.HandleFunc("/wallet/vcs/", handleWalletShow)                // GET:  show raw VC by ID
	mux.HandleFunc("/wallet/verify", handleWalletVerify)            // POST: verify a VC (sig, expiry, etc.)
	mux.HandleFunc("/wallet/vcs-pretty/", handleWalletClaims)       // GET:  show parsed claims for a VC
	mux.HandleFunc("/wallet/ingest", handleWalletIngest)            // POST: ingest a compact JWS directly (new)

	// Start server
	log.Printf("Listening on :8080 (public origin: %s)", issuerBaseURL)
	if err := http.ListenAndServe(":8080", withCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

// Small helper so the compiler pulls in “strings” used by main package only here.
var _ = strings.Join
var _ = time.Now
