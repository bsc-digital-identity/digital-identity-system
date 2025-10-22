package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"log"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

// generateIssuerKey generates a new Ed25519 key pair for the issuer (this server).
// The private key is stored globally for signing credentials (VCs, JWTs, etc.).
// The public key is published in /.well-known/jwks.json so verifiers can validate signatures.
func generateIssuerKey() {
	// --- 1. Generate a new Ed25519 key pair ---
	// Ed25519 provides fast and secure digital signatures with small key size.
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatalf("keygen: %v", err)
	}
	issuerPrivKey = priv // store private key globally for signing JWTs

	// --- 2. Convert the public key into JWK (JSON Web Key) format ---
	k, err := jwk.FromRaw(pub)
	if err != nil {
		log.Fatalf("jwk from raw: %v", err)
	}

	// --- 3. Assign a unique Key ID (kid) ---
	// This allows verifiers to identify which key was used to sign a token.
	if err := jwk.AssignKeyID(k); err != nil {
		log.Fatalf("assign kid: %v", err)
	}
	issuerKeyID = k.KeyID() // save the generated key ID for JWT headers

	// --- 4. Create a JWK Set containing this public key ---
	// This set will later be exposed via /.well-known/jwks.json
	issuerJWKSet = jwk.NewSet()
	issuerJWKSet.AddKey(k)
}
