package server

import (
	"errors"
	"net/http"
	"zk-wallet-go/internal/app/config"
)

// CurrentSession retrieves the current user session based on the session cookie.
//
// It looks up the "sid" cookie (as defined in cookieNameSession), and then checks
// the in-memory session store (sessStore) for the corresponding session record.
//
// Returns:
//   - (session, nil) if a valid session is found
//   - (empty session, error) if missing or invalid cookie / session not found
//
// NOTE:
//
//	This is an in-memory session helper only suitable for demos / PoC.
//	In production, replace with persistent session storage (e.g., Redis or DB).
func CurrentSession(r *http.Request) (config.Session, error) {
	// Try to read the session cookie from the incoming request.
	c, _ := r.Cookie(config.CookieNameSession)
	if c == nil {
		return config.Session{}, errors.New("no cookie") // user is not logged in
	}

	// Look up session by ID in the in-memory store (thread-safe read).
	config.SessMu.RLock()
	s, ok := config.SessStore[c.Value]
	config.SessMu.RUnlock()

	if !ok {
		return config.Session{}, errors.New("not found") // cookie present but session expired/invalid
	}

	// Session found — return it to the caller (e.g., handler)
	return s, nil
}
