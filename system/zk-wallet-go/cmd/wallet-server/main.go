package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"

	"zk-wallet-go/internal/app/config"
	"zk-wallet-go/internal/app/handlers"
	"zk-wallet-go/internal/app/keys"
	"zk-wallet-go/internal/app/oidcissuer"
	"zk-wallet-go/internal/app/zkprequest"
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
	config.RedirectURI = config.GetenvDefault(
		"OIDC_REDIRECT_URI",
		config.IssuerBaseURL+"/auth/dsnet/callback",
	)
	// external ZKP verifier (presentation request service, e.g. :9000)
	config.ZkpVerifierBaseURL = config.MustEnv("ZKP_VERIFIER_BASE_URL")

	// Read PORT from env
	port := config.MustEnv("PORT")

	// --- OIDC discovery (DSNet as OP) ---
	var err error
	config.OidcProvider, err = oidc.NewProvider(
		context.Background(),
		config.DsnetIssuer,
	)
	if err != nil {
		log.Fatalf("OIDC discovery failed: %v", err)
	}

	config.Verifier = config.OidcProvider.Verifier(&oidc.Config{
		ClientID: config.OidcClientID,
	})

	config.Oauth2Config = &oauth2.Config{
		ClientID:     config.OidcClientID,
		ClientSecret: config.OidcClientSecret,
		Endpoint:     config.OidcProvider.Endpoint(),
		Scopes:       []string{"openid", "email", "profile"},
		RedirectURL:  config.RedirectURI,
	}

	log.Printf("[CONFIG] DSNET issuer:        %s", config.DsnetIssuer)
	log.Printf("[CONFIG] Redirect URI:        %s", config.RedirectURI)
	log.Printf("[CONFIG] Client ID:           %s", config.OidcClientID)
	log.Printf("[CONFIG] Using secret?        %v", config.OidcClientSecret != "")
	log.Printf("[CONFIG] Token endpoint:      %s", config.Oauth2Config.Endpoint.TokenURL)
	log.Printf("[CONFIG] ZKP verifier base:   %s", config.ZkpVerifierBaseURL)
	log.Printf("[CONFIG] Server port:         %s", port)

	// --- Issuer keypair + JWKS ---
	keys.GenerateIssuerKey()

	// --- Router setup ---
	mux := http.NewServeMux()

	// Static assets (wallet UI / demo pages)
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

	// --- Wallet MVP endpoints (VC storage + verification) ---
	mux.HandleFunc("/wallet/import-offer", handlers.HandleWalletImportOffer)
	mux.HandleFunc("/wallet/vcs", handlers.HandleWalletList)
	mux.HandleFunc("/wallet/vcs/", handlers.HandleWalletShow)
	mux.HandleFunc("/wallet/verify", handlers.HandleWalletVerify)
	mux.HandleFunc("/wallet/vcs-pretty/", handlers.HandleWalletClaims)
	mux.HandleFunc("/wallet/ingest", handlers.HandleWalletIngest)

	// --- Local ZKP Presentation Request service (demo RP) ---
	// To jest lokalny RP z domyślnym schematem age_over_18; niezależny od dynamicznego schematu
	// używanego przez zewnętrzny verifier na :9000.
	zkpSvc := zkprequest.NewService(
		&zkprequest.InMemoryStore{},
		&zkprequest.StaticRegistry{
			Schemas: map[string]string{},
		},
		func(s *zkprequest.Service) {
			s.Audience = config.IssuerBaseURL
			s.ResponseURI = config.IssuerBaseURL + "/zkp/verify"
			s.TTL = 15 * time.Minute
		},
	)

	zkpHandlers := &zkprequest.Handlers{Svc: zkpSvc}

	// Local presentation request endpoints (demo)
	mux.HandleFunc("/zkp/request", zkpHandlers.Request)
	mux.HandleFunc("/zkp/verify", zkpHandlers.Verify)

	// --- Dynamic ZKP integration with external verifier ---
	// 1) Given request_id -> fetch descriptor & schema from {ZKP_VERIFIER_BASE_URL}
	mux.HandleFunc("/wallet/zkp/fetch-descriptor", handlers.HandleFetchDescriptor)
	// 2) Given request_id + inputs -> fetch schema, build circuit, create ZKP
	mux.HandleFunc("/wallet/zkp/create", handlers.HandleZkpCreate)
	// 3) Dev-only endpoint to "force-verify" proof on external verifier
	//mux.HandleFunc("/wallet/zkp/verify", handlers.HandleVerifyProxy)

	// --- Start HTTP server ---
	log.Printf("Listening on :%s (public origin: %s)", port, config.IssuerBaseURL)
	if err := http.ListenAndServe(":"+port, util.WithCORS(mux)); err != nil {
		log.Fatal(err)
	}
}
