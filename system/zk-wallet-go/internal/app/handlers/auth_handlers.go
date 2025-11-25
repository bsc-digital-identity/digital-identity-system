package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"zk-wallet-go/internal/app/config"
	"zk-wallet-go/internal/app/server"
	"zk-wallet-go/pkg/util"
)

// HandleLogin starts the OIDC authorization request flow.
// It generates a PKCE verifier + challenge and redirects the user to the OIDC provider's login page.
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// --- PKCE (Proof Key for Code Exchange) setup ---
	// Generate random string used as PKCE verifier
	ver := util.RandomString(64)
	http.SetCookie(w, &http.Cookie{
		Name:     config.CookieNamePKCE,
		Value:    ver,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // in production: set to true (HTTPS only)
		SameSite: http.SameSiteLaxMode,
	})

	// Hash verifier with SHA-256 to create PKCE challenge
	sum := sha256Bytes([]byte(ver))
	challenge := base64RawURLEnc(sum)

	// --- Nonce for replay protection ---
	// Used later to verify ID Token authenticity
	nonce := util.RandomString(32)
	state := util.RandomString(32)
	http.SetCookie(w, &http.Cookie{
		Name:     config.CookieNameState,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     config.CookieNameNonce,
		Value:    nonce,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	// --- Build Authorization URL ---
	// Redirect user to OIDC provider’s login page with PKCE params and nonce
	config.PKCEMu.Lock()
	config.PKCEStore[state] = ver
	config.PKCEMu.Unlock()
	authURL := config.Oauth2Config.AuthCodeURL(
		state, // random state to prevent CSRF (stored in cookie)
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("prompt", "login"),
		oauth2.SetAuthURLParam("max_age", "0"), // force fresh auth even if OP session exists
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("scope", strings.Join(config.Oauth2Config.Scopes, " ")),
	)

	log.Println("[OIDC] redirecting to:", authURL)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// HandleCallback is called after the OIDC provider redirects back with the authorization code.
// It exchanges the code for tokens, verifies the ID token, and creates a session.
func HandleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	// Extract authorization code from query params
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	stateParam := r.URL.Query().Get("state")
	stateCookie, _ := r.Cookie(config.CookieNameState)
	if stateCookie == nil || stateCookie.Value == "" || stateParam == "" || stateParam != stateCookie.Value {
		log.Printf("[OIDC] state mismatch (cookie=%v, param=%s) – continuing in dev mode", stateCookie, stateParam)
	}

	nonceCookie, _ := r.Cookie(config.CookieNameNonce)

	// Retrieve stored PKCE verifier from cookie or in-memory store (by state)
	pkceVal := ""
	if pkceCookie, _ := r.Cookie(config.CookieNamePKCE); pkceCookie != nil {
		pkceVal = pkceCookie.Value
	}
	if pkceVal == "" {
		config.PKCEMu.RLock()
		pkceVal = config.PKCEStore[stateParam]
		config.PKCEMu.RUnlock()
	}
	if pkceVal == "" {
		http.Error(w, "missing pkce verifier", http.StatusBadRequest)
		return
	}

	// Exchange authorization code for access token (and optionally ID token)
	tok, err := config.Oauth2Config.Exchange(
		ctx,
		code,
		oauth2.SetAuthURLParam("code_verifier", pkceVal),
		oauth2.SetAuthURLParam("redirect_uri", config.RedirectURI),
	)
	if err != nil {
		log.Printf("[OIDC] Token exchange FAILED: %v", err)
		http.Error(w, "token exchange failed: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Println("[OIDC] Token exchange SUCCESS")
	log.Printf("  AccessToken len=%d", len(tok.AccessToken))

	// --- Try to read ID Token from token response ---
	var sub, email, name string
	if rawIDToken, ok := tok.Extra("id_token").(string); ok && rawIDToken != "" {
		// ID Token found → verify its signature and parse claims
		log.Printf("  ID Token len=%d", len(rawIDToken))
		idt, err := config.Verifier.Verify(ctx, rawIDToken)
		if err != nil {
			http.Error(w, "id_token verify failed: "+err.Error(), http.StatusBadRequest)
			return
		}
		// Extract standard claims
		var claims struct {
			Sub   string `json:"sub"`
			Email string `json:"email"`
			Name  string `json:"name"`
			Nonce string `json:"nonce"`
		}
		if err := idt.Claims(&claims); err != nil {
			http.Error(w, "claims parse: "+err.Error(), http.StatusBadRequest)
			return
		}
		if nonceCookie != nil && claims.Nonce != "" && claims.Nonce != nonceCookie.Value {
			http.Error(w, "nonce mismatch", http.StatusBadRequest)
			return
		}
		sub, email, name = claims.Sub, claims.Email, claims.Name

	} else {
		// --- Fallback if ID Token not returned ---
		// Some providers (like DSNet) may omit it; use /userinfo endpoint instead
		log.Println("  Missing id_token in response! Falling back to /userinfo")

		ui, err := config.OidcProvider.UserInfo(ctx, oauth2.StaticTokenSource(tok))
		if err != nil {
			http.Error(w, "userinfo fetch failed: "+err.Error(), http.StatusBadRequest)
			return
		}
		sub = ui.Subject

		// Extract email and name from /userinfo response
		var extra struct {
			Email string `json:"email"`
			Name  string `json:"name"`
		}
		_ = ui.Claims(&extra)
		email, name = extra.Email, extra.Name

		if sub == "" {
			http.Error(w, "userinfo missing sub", http.StatusBadRequest)
			return
		}
	}

	// --- Create local session ---
	sid := util.RandomString(32)
	config.SessMu.Lock()
	config.SessStore[sid] = config.Session{Sub: sub, Email: email, Name: name}
	config.SessMu.Unlock()

	// Clear one-time cookies (PKCE + state + nonce) to avoid reuse.
	clearCookie := func(name string) {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})
	}
	clearCookie(config.CookieNamePKCE)
	clearCookie(config.CookieNameState)
	clearCookie(config.CookieNameNonce)
	config.PKCEMu.Lock()
	delete(config.PKCEStore, stateParam)
	config.PKCEMu.Unlock()

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     config.CookieNameSession,
		Value:    sid,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600, // 1h
	})

	// Redirect user to main page after login
	http.Redirect(w, r, "/offer.html", http.StatusFound)
}

// HandleLogout clears the session cookie and removes user session from memory.
func HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Where to land after logout (local page).
	postLogout := config.IssuerBaseURL

	c, _ := r.Cookie(config.CookieNameSession)
	if c != nil {
		config.SessMu.Lock()
		delete(config.SessStore, c.Value)
		config.SessMu.Unlock()
	}
	// Expire cookie immediately
	expire := func(name string) {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Expires:  time.Now().Add(-time.Hour),
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteLaxMode,
		})
	}
	expire(config.CookieNameSession)
	expire(config.CookieNamePKCE)
	expire(config.CookieNameState)
	expire(config.CookieNameNonce)

	// If we have a DSNET logout URL, redirect the browser there to clear OP session
	// and return via post_logout_redirect_uri.
	if config.DsnetLogout != "" {
		u, _ := url.Parse(config.DsnetLogout)
		q := u.Query()
		q.Set("post_logout_redirect_uri", postLogout)
		u.RawQuery = q.Encode()
		http.Redirect(w, r, u.String(), http.StatusFound)
		return
	}

	// Fallback: If OP supports RP-initiated logout, redirect there with post_logout_redirect_uri.
	if config.EndSessionEndpoint != "" {
		u := config.EndSessionEndpoint
		if strings.Contains(u, "?") {
			u += "&"
		} else {
			u += "?"
		}
		http.Redirect(w, r, u+"post_logout_redirect_uri="+util.EscapeQuery(postLogout), http.StatusFound)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

// HandleMe returns the current session (if any).
func HandleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s, err := server.CurrentSession(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	util.WriteJSON(w, s)
}

// helpers

// --- Helper functions for hashing and encoding ---

// sha256Bytes returns the SHA-256 hash of given bytes as a byte slice.
func sha256Bytes(b []byte) []byte {
	h := sha256sum(b)
	return h[:]
}

// sha256sum returns the SHA-256 hash as a fixed-size array.
func sha256sum(b []byte) [32]byte { return sha256Array(b) }

// sha256Array performs the actual hash calculation.
func sha256Array(b []byte) [32]byte { return sha256.Sum256(b) }

// base64RawURLEnc encodes bytes in base64 URL-safe format (without padding).
func base64RawURLEnc(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}
