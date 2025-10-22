package main

import (
	"encoding/json"
	"net/http"
)

// handleCredentialIssuerMetadata serves OIDC4VCI "Credential Issuer Metadata".
// This advertises your issuer/AS endpoints and the set of credentials you can issue.
// Spec: OpenID for Verifiable Credential Issuance (OIDC4VCI).
func handleCredentialIssuerMetadata(w http.ResponseWriter, r *http.Request) {
	base := issuerBaseURL

	// Minimal-but-useful metadata. For a colocated Authorization Server, we point
	// both "credential_issuer" and "authorization_server" to the same base.
	// NOTE: In a multi-component deployment, "authorization_server" may be a different host.
	meta := map[string]any{
		// Identifier (URL) of this credential issuer
		"credential_issuer": base,

		// OAuth2/OIDC Authorization Server used for token/authorization
		"authorization_server": base, // colocated AS

		// OAuth2 token endpoint for OIDC4VCI token requests
		"token_endpoint": base + "/oidc4vci/token",

		// OIDC4VCI credential endpoint (wallet hits this with proof to obtain VC)
		"credential_endpoint": base + "/oidc4vci/credential",

		// Declares which credentials can be issued and how
		"credentials_supported": []any{
			map[string]any{
				// Local identifier for this credential configuration
				"credential_configuration_id": "StudentCredential_JWT_v1",

				// Using compact JWS-based VC
				"format": "jwt_vc_json",

				// OAuth2 scope to request this credential (wallet will ask for it)
				"scope": "student",

				// VC schema definition (types + subject fields)
				"credential_definition": map[string]any{
					"type": []string{"VerifiableCredential", "StudentCredential"},
					"credentialSubject": map[string]any{
						"age": map[string]any{
							"display": []any{
								map[string]any{"name": "Age", "locale": "en"},
							},
						},
						"student_status": map[string]any{
							"display": []any{
								map[string]any{"name": "Student Status", "locale": "en"},
							},
						},
						"university": map[string]any{
							"display": []any{
								map[string]any{"name": "University", "locale": "en"},
							},
						},
						"birthdate": map[string]any{
							"display": []any{
								map[string]any{"name": "Date of Birth", "locale": "en"},
							},
						},
					},
				},

				// How the holder binds their key to the request (JWK thumbprint etc.)
				"cryptographic_binding_methods_supported": []string{"jwk"},

				// Proof-of-possession formats and supported algs (for /credential)
				"proof_types_supported": map[string]any{
					"jwt": map[string]any{
						"alg_values_supported": []string{"EdDSA", "ES256"},
					},
				},

				// UX-facing display metadata for wallets
				"display": []any{
					map[string]any{
						"name":        "AGH Student Credential",
						"description": "Official credential confirming active student status at AGH University.",
						"locale":      "en",
					},
				},
			},
		},
	}

	// Writes pretty JSON with correct Content-Type
	writeJSON(w, meta)
}

// handleJWKS exposes your issuer's public keys in JWKS format.
// Wallets/verifiers use this to verify signatures produced by your private key.
// TIP: Consider setting Cache-Control/ETag for key rotation strategies.
func handleJWKS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(issuerJWKSet)
}
