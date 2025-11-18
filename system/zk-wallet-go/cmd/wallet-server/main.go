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
	"zk-wallet-go/internal/app/vcstore"
	"zk-wallet-go/internal/app/zkprequest"
	"zk-wallet-go/pkg/util"
)

func main() {
	// Wczytanie .env (dla uruchamiania lokalnie)
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
	config.ZkpVerifierBaseURL = config.MustEnv("ZKP_VERIFIER_BASE_URL")

	port := config.MustEnv("PORT")

	// --- OIDC discovery (DSNet jako OP) ---
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

	// --- VCStore + Handlers ---
	vcStore := vcstore.NewInMemoryVCStore()

	// portfel (import, list, show, verify, ingest) oparty na VCStore
	walletHandler := handlers.NewWalletHandler(vcStore)

	// ZKP handler, który bierze dane do dowodu z VCStore
	zkpWalletHandler := handlers.NewZkpHandler(vcStore)

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

	// --- Wallet endpoints oparte na VCStore ---
	// import-offer starego typu (tylko offer) możesz zostawić lub docelowo zastąpić ingestem
	mux.HandleFunc("/wallet/import-offer", walletHandler.HandleWalletImportOffer)

	// unified ingest (credential OR offer) – MUSI być metodą na WalletHandler,
	// przerobioną tak, żeby używała vcStore.Save(...)
	mux.HandleFunc("/wallet/ingest", walletHandler.HandleWalletIngest)

	// lista / pojedynczy VC / verify – przez VCStore
	mux.HandleFunc("/wallet/vcs", walletHandler.HandleWalletList)
	mux.HandleFunc("/wallet/vcs/", walletHandler.HandleWalletShow)
	mux.HandleFunc("/wallet/verify", walletHandler.HandleWalletVerify)

	// pretty claims do debugowania → ZkpHandler oparty na VCStore
	mux.HandleFunc("/wallet/vcs-pretty/", zkpWalletHandler.HandleWalletClaims)

	// --- Local ZKP Presentation Request service (demo RP) ---
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

	mux.HandleFunc("/zkp/request", zkpHandlers.Request)
	mux.HandleFunc("/zkp/verify", zkpHandlers.Verify)

	// --- Dynamic ZKP integration with external verifier ---
	mux.HandleFunc("/wallet/zkp/fetch-descriptor", handlers.HandleFetchDescriptor)
	mux.HandleFunc("/wallet/zkp/create", zkpWalletHandler.HandleZkpCreate)

	log.Printf("Listening on :%s (public origin: %s)", port, config.IssuerBaseURL)
	if err := http.ListenAndServe(":"+port, util.WithCORS(mux)); err != nil {
		log.Fatal(err)
	}
}
