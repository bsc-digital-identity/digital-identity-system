package main

import (
	"errors"
	"net/http"
)

// currentSession retrieves the current user session based on the session cookie.
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
func currentSession(r *http.Request) (session, error) {
	// Try to read the session cookie from the incoming request.
	c, _ := r.Cookie(cookieNameSession)
	if c == nil {
		return session{}, errors.New("no cookie") // user is not logged in
	}

	// Look up session by ID in the in-memory store (thread-safe read).
	sessMu.RLock()
	s, ok := sessStore[c.Value]
	sessMu.RUnlock()

	if !ok {
		return session{}, errors.New("not found") // cookie present but session expired/invalid
	}

	// Session found — return it to the caller (e.g., handler)
	return s, nil
}
