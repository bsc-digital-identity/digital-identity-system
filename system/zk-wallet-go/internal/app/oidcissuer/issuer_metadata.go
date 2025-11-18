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
				// Local identifier for this ZK config / credential type
				"credential_configuration_id": "AgeOver18_JWT_v1",

				// VC format
				"format": "jwt_vc_json",

				// OIDC scope the wallet requests to obtain this VC
				"scope": "age_over_18",

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

				// VC definition aligned with the ZK schema: birth_ts is the secret field in the circuit,
				// but still a claim in the credential itself.
				"credential_definition": map[string]any{
					"type": []string{"VerifiableCredential", "AgeOver18Credential"},
					"credentialSubject": map[string]any{
						"birth_ts": map[string]any{
							"display": []any{
								map[string]any{
									"name":   "Birth timestamp (UNIX seconds)",
									"locale": "en",
								},
							},
						},
					},
				},

				// Wallet UX labels
				"display": []any{
					map[string]any{
						"name":        "Age over 18 credential",
						"description": "Credential containing birth timestamp to prove that the holder is at least 18 years old via zero-knowledge proof.",
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
