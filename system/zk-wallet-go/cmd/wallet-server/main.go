package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"
	"zk-wallet-go/internal/app/zkp"
	"zk-wallet-go/internal/app/zkprequest"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"

	"zk-wallet-go/internal/app/config"
	"zk-wallet-go/internal/app/handlers"
	"zk-wallet-go/internal/app/keys"
	"zk-wallet-go/internal/app/oidcissuer"
	"zk-wallet-go/pkg/util"
)

func main() {
	// Load .env file (works when running from cmd/wallet-server)
	_ = godotenv.Load()

	// --- Environment configuration ---
	config.IssuerBaseURL = config.MustEnv("ISSUER_BASE_URL")
	config.DsnetIssuer = config.MustEnv("DSNET_ISSUER")
	config.OidcClientID = config.MustEnv("OIDC_CLIENT_ID")
	config.OidcClientSecret = config.MustEnv("OIDC_CLIENT_SECRET")
	config.RedirectURI = config.GetenvDefault("OIDC_REDIRECT_URI", config.IssuerBaseURL+"/auth/dsnet/callback")

	// Read PORT from env (default to 8080)
	port := config.MustEnv("PORT")

	// --- OIDC discovery ---
	var err error
	config.OidcProvider, err = oidc.NewProvider(context.Background(), config.DsnetIssuer)
	if err != nil {
		log.Fatalf("OIDC discover failed: %v", err)
	}

	config.Verifier = config.OidcProvider.Verifier(&oidc.Config{ClientID: config.OidcClientID})
	config.Oauth2Config = &oauth2.Config{
		ClientID:     config.OidcClientID,
		ClientSecret: config.OidcClientSecret,
		Endpoint:     config.OidcProvider.Endpoint(),
		Scopes:       []string{"openid", "email", "profile"},
		RedirectURL:  config.RedirectURI,
	}

	log.Printf("[CONFIG] DSNET issuer: %s", config.DsnetIssuer)
	log.Printf("[CONFIG] Redirect URI: %s", config.RedirectURI)
	log.Printf("[CONFIG] Client ID: %s", config.OidcClientID)
	log.Printf("[CONFIG] Using secret? %v", config.OidcClientSecret != "")
	log.Printf("[CONFIG] Token endpoint: %s", config.Oauth2Config.Endpoint.TokenURL)
	log.Printf("[CONFIG] Server port: %s", port)

	// --- Issuer keypair + JWKS ---
	keys.GenerateIssuerKey()

	// --- Router setup ---
	mux := http.NewServeMux()

	// Static assets
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/", fs)

	// --- DSNET login (OIDC code flow with PKCE) ---
	mux.HandleFunc("/auth/dsnet/login", handlers.HandleLogin)
	mux.HandleFunc("/auth/dsnet/callback", handlers.HandleCallback)
	mux.HandleFunc("/auth/logout", handlers.HandleLogout)

	// --- OIDC4VCI issuer metadata & JWKS ---
	mux.HandleFunc("/.well-known/openid-credential-issuer", oidcissuer.HandleCredentialIssuerMetadata)
	mux.HandleFunc("/.well-known/jwks.json", oidcissuer.HandleJWKS)

	// --- OIDC4VCI protocol endpoints ---
	mux.HandleFunc("/oidc4vci/offer", handlers.HandleCreateOffer)
	mux.HandleFunc("/oidc4vci/offer/", handlers.HandleOfferByCode)
	mux.HandleFunc("/oidc4vci/token", handlers.HandleVciToken)
	mux.HandleFunc("/oidc4vci/credential", handlers.HandleVciCredential)

	// --- Wallet MVP endpoints ---
	mux.HandleFunc("/wallet/import-offer", handlers.HandleWalletImportOffer)
	mux.HandleFunc("/wallet/vcs", handlers.HandleWalletList)
	mux.HandleFunc("/wallet/vcs/", handlers.HandleWalletShow)
	mux.HandleFunc("/wallet/verify", handlers.HandleWalletVerify)
	mux.HandleFunc("/wallet/vcs-pretty/", handlers.HandleWalletClaims)
	mux.HandleFunc("/wallet/ingest", handlers.HandleWalletIngest)

	// --- ZKP Presentation Request service ---
	zkpSvc := &zkprequest.Service{
		Store: &zkprequest.InMemoryStore{},
		Circuits: &zkprequest.StaticRegistry{
			Schemas: map[string]string{
				"age_over_18@1": zkp.DefaultAgeSchema, // demo circuit
			},
			// VKs: map[string][]byte{ "age_over_18@1": <server-held-verifying-key-bytes> }, // prod
		},
		Audience:           config.IssuerBaseURL,                 // who the proof is for (your verifier origin)
		ResponseURI:        config.IssuerBaseURL + "/zkp/verify", // where wallets POST proofs
		TTL:                5 * time.Minute,                      // request validity
		VerifyWithServerVK: false,                                // set true in prod when you hold VK server-side
	}

	zkpHandlers := &zkprequest.Handlers{Svc: zkpSvc}

	// Expose verifier endpoints
	mux.HandleFunc("/zkp/request", zkpHandlers.Request)
	mux.HandleFunc("/zkp/verify", zkpHandlers.Verify)

	// --- Start HTTP server ---
	log.Printf("Listening on :%s (public origin: %s)", port, config.IssuerBaseURL)
	if err := http.ListenAndServe(":"+port, util.WithCORS(mux)); err != nil {
		log.Fatal(err)
	}
}

// Keep these references to avoid unused imports
var _ = strings.Join
var _ = time.Now
