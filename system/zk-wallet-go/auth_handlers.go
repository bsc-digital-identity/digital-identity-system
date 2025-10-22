package main

import (
	"crypto/sha256"
	"encoding/base64"
	"log"
	"net/http"
	"strings"

	"golang.org/x/oauth2"
)

// handleLogin starts the OIDC authorization request flow.
// It generates a PKCE verifier + challenge and redirects the user to the OIDC provider's login page.
func handleLogin(w http.ResponseWriter, r *http.Request) {
	// --- PKCE (Proof Key for Code Exchange) setup ---
	// Generate random string used as PKCE verifier
	ver := randomString(64)
	http.SetCookie(w, &http.Cookie{
		Name:     cookieNamePKCE,
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
	nonce := randomString(16)
	http.SetCookie(w, &http.Cookie{
		Name:     "oidc_nonce",
		Value:    nonce,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	// --- Build Authorization URL ---
	// Redirect user to OIDC provider’s login page with PKCE params and nonce
	authURL := oauth2Config.AuthCodeURL(
		randomString(16), // random state to prevent CSRF
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("prompt", "login"),
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("scope", strings.Join(oauth2Config.Scopes, " ")),
	)

	log.Println("[OIDC] redirecting to:", authURL)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// handleCallback is called after the OIDC provider redirects back with the authorization code.
// It exchanges the code for tokens, verifies the ID token, and creates a session.
func handleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract authorization code from query params
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	// Retrieve stored PKCE verifier from cookie
	pkce, _ := r.Cookie(cookieNamePKCE)
	if pkce == nil {
		http.Error(w, "missing pkce verifier", http.StatusBadRequest)
		return
	}

	// Exchange authorization code for access token (and optionally ID token)
	tok, err := oauth2Config.Exchange(
		ctx,
		code,
		oauth2.SetAuthURLParam("code_verifier", pkce.Value),
		oauth2.SetAuthURLParam("redirect_uri", redirectURI),
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
		idt, err := verifier.Verify(ctx, rawIDToken)
		if err != nil {
			http.Error(w, "id_token verify failed: "+err.Error(), http.StatusBadRequest)
			return
		}
		// Extract standard claims
		var claims struct {
			Sub   string `json:"sub"`
			Email string `json:"email"`
			Name  string `json:"name"`
		}
		if err := idt.Claims(&claims); err != nil {
			http.Error(w, "claims parse: "+err.Error(), http.StatusBadRequest)
			return
		}
		sub, email, name = claims.Sub, claims.Email, claims.Name

	} else {
		// --- Fallback if ID Token not returned ---
		// Some providers (like DSNet) may omit it; use /userinfo endpoint instead
		log.Println("  Missing id_token in response! Falling back to /userinfo")

		ui, err := oidcProvider.UserInfo(ctx, oauth2.StaticTokenSource(tok))
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
	sid := randomString(32)
	sessMu.Lock()
	sessStore[sid] = session{Sub: sub, Email: email, Name: name}
	sessMu.Unlock()

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     cookieNameSession,
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

// handleLogout clears the session cookie and removes user session from memory.
func handleLogout(w http.ResponseWriter, r *http.Request) {
	c, _ := r.Cookie(cookieNameSession)
	if c != nil {
		sessMu.Lock()
		delete(sessStore, c.Value)
		sessMu.Unlock()
	}
	// Expire cookie immediately
	http.SetCookie(w, &http.Cookie{Name: cookieNameSession, Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/", http.StatusFound)
}

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
