package oidcissuer

import (
	"encoding/json"
	"net/http"

	"zk-wallet-go/internal/app/config"
	"zk-wallet-go/pkg/util"
)

// HandleCredentialIssuerMetadata serves OIDC4VCI "Credential Issuer Metadata".
// Spec-aligned for jwt_vc_json, PoP via cnf.jwk, and pre-authorized_code flow.
func HandleCredentialIssuerMetadata(w http.ResponseWriter, r *http.Request) {
	base := config.IssuerBaseURL

	meta := map[string]any{
		// REQUIRED: identifier (URL) of this credential issuer
		"credential_issuer": base,

		// If the AS is co-located, you can point to the same origin
		"authorization_server": base,

		// Endpoints used by wallets
		"token_endpoint":      base + "/oidc4vci/token",
		"credential_endpoint": base + "/oidc4vci/credential",

		// Grants the issuer supports (auth code + pre-authorized)
		"grants": map[string]any{
			"authorization_code": map[string]any{},
			"urn:ietf:params:oauth:grant-type:pre-authorized_code": map[string]any{
				"tx_code": map[string]any{
					"length":     8,
					"input_mode": "numeric",
				},
			},
		},

		// What credentials this issuer can issue
		"credentials_supported": []any{
			map[string]any{
				// Local identifier for this zkpconfig (wallet requests this via scope/metadata)
				"credential_configuration_id": "StudentCredential_JWT_v1",

				// VC format
				"format": "jwt_vc_json",

				// OIDC scope the wallet requests to obtain this VC
				"scope": "student",

				// Which algs the issuer uses to sign the VC (the JWT)
				"credential_signing_alg_values_supported": []string{"EdDSA", "ES256"},

				// Proof-of-possession required at /credential (holder binding)
				"cryptographic_binding_methods_supported": []string{"jwk"},
				"proof_types_supported": map[string]any{
					"jwt": map[string]any{
						// algs supported for the holder's proof JWT at /credential
						"alg_values_supported": []string{"EdDSA", "ES256"},
					},
				},

				// The VC's high-level definition (types + subject claim descriptions)
				"credential_definition": map[string]any{
					"type": []string{"VerifiableCredential", "StudentCredential"},
					"credentialSubject": map[string]any{
						"age": map[string]any{
							"display": []any{
								map[string]any{"name": "Age", "locale": "en"},
							},
						},
						"university": map[string]any{
							"display": []any{
								map[string]any{"name": "University", "locale": "en"},
							},
						},
						"birthDate": map[string]any{
							"display": []any{
								map[string]any{"name": "Date of Birth", "locale": "en"},
							},
						},
						"student_id": map[string]any{
							"display": []any{
								map[string]any{"name": "Student ID", "locale": "en"},
							},
						},
						"student_status": map[string]any{
							"display": []any{
								map[string]any{"name": "Student Status", "locale": "en"},
							},
						},
						"gender": map[string]any{
							"display": []any{
								map[string]any{"name": "Gender", "locale": "en"},
							},
						},
					},
				},

				// Wallet UX labels
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

	// Optional niceties (helps wallets discover keys):
	// meta["jwks_uri"] = base + "/.well-known/jwks.json"

	util.WriteJSON(w, meta)
}

// HandleJWKS exposes issuer public keys (for verifying signed JWT-VCs).
func HandleJWKS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(config.IssuerJWKSet)
}
