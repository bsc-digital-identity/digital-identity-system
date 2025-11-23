# 🔐 zk-wallet-go
### Minimal OIDC4VCI Issuer + Wallet MVP (DSNET SSO + Verifiable Credentials)

**zk-wallet-go** is a minimal proof-of-concept implementation of **OpenID Connect for Verifiable Credential Issuance (OIDC4VCI)** in Go, integrated with **DSNET SSO (OIDC PKCE)**.  
It acts as both a local **issuer** and a simple **wallet MVP**, allowing users to log in via DSNET and issue or store Verifiable Credentials (VC).

---

## 🧩 Features

- 🔑 **User login via DSNET** (OIDC Authorization Code + PKCE)
- 🪪 **OIDC4VCI endpoints:**
  - `/.well-known/openid-credential-issuer`
  - `/oidc4vci/offer`
  - `/oidc4vci/token`
  - `/oidc4vci/credential`
- 💾 **Wallet MVP**:
  - `/wallet/ingest` — fetches and stores a credential from an offer
  - `/wallet/vcs-pretty/<VC_ID>` — displays stored credentials in readable JSON
- 🌐 Simple `offer.html` page generating credential offers and deeplinks

---

## ⚙️ Quick Start

### Clone & configure

```bash
git clone https://github.com/<your-org>/zk-wallet-go.git
cd zk-wallet-go
Create .env file in zk-wallet-go directory
Edit .env to set your DSNET SSO client ID/secret and redirect URI:

ISSUER_BASE_URL=http://localhost:8080
DSNET_ISSUER=<your-dsnet-issuer-url>
OIDC_CLIENT_ID=<your-dsnet-client-id>
OIDC_CLIENT_SECRET=<your-dsnet-client-secret>
OIDC_REDIRECT_URI=http://localhost:8080/auth/dsnet/callback
```

### Run locally

```bash
go mod init zk-wallet-go
go get github.com/coreos/go-oidc/v3/oidc \
      github.com/lestrrat-go/jwx/v2/jws \
      github.com/lestrrat-go/jwx/v2/jwk \
      github.com/joho/godotenv \
      golang.org/x/oauth2


0. go run .
1. localhost:8080
2. Login via DSNET
3. Create offer
4. C:\Windows\System32>curl -i -X POST "http://localhost:8080/wallet/ingest" ^
More?   -H "Content-Type: application/json" ^
More?   -d "{\"offer\":\"http://localhost:8080/oidc4vci/offer/<OFFER_ID>\"}"

Example response:
HTTP/1.1 200 OK
Access-Control-Allow-Credentials: true
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Allow-Methods: GET,POST,OPTIONS
Access-Control-Allow-Origin: http://localhost:8080
Content-Type: application/json
Vary: Origin
Date: Wed, 22 Oct 2025 10:43:40 GMT
Content-Length: 118

{
  "source": "offer",
  "valid": false,
  "vc_id": "<VC_ID>",
  "stored": true
}

5. C:\Windows\System32>curl "http://localhost:8080/wallet/vcs-pretty/<VC_ID>"
Example response:
{
  "claims": {
    "iat": 1761129820,
    "iss": "http://localhost:8080",
    "nbf": 1761129820,
    "sub": "t_83172",
    "vc": {
      "@context": [
        "https://www.w3.org/2018/credentials/v1"
      ],
      "credentialSubject": {
        "birthdate": "2002-05-10",
        "gender": "unspecified",
        "id": "t_83172",
        "student_id": "123456",
        "student_status": "active"
      },
      "issuanceDate": "2025-10-22T10:43:40Z",
      "issuer": "http://localhost:8080",
      "type": [
        "VerifiableCredential",
        "StudentCredential"
      ]
    }
  },
  "display_name": "Credential",
  "format": "jwt_vc_json",
  "id": "<VC_ID>",
  "issuer_meta": "",
  "subject_meta": "",
  "types_meta": null
}
```

```bash
go run cmd/wallet-server/main.go
# localhost:8087 -> login, create offer
curl -i -X POST "http://localhost:8087/wallet/ingest" -H "Content-Type: application/json" -d "{\"offer\":\"http://localhost:8087/oidc4vci/offer/<id>\"}"
curl "http://localhost:8087/wallet/vcs-pretty/<vc_id>"
```
