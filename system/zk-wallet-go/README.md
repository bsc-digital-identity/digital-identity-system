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

- Register on [DSNET](https://akademik.agh.edu.pl/auth/register)
- Generate new app in [DSNET](https://panel.dsnet.agh.edu.pl/)
- Set `OIDC_CLIENT_ID` and `OIDC_CLIENT_SECRET` in `.env`
- `go run cmd/wallet-server/main.go`
- Open `http://localhost:8087/` in browser and login via DSNET
- Open `http://localhost:8087/offer.html` and generate a credential offer
- `curl -i -X POST 'http://localhost:8087/wallet/ingest' -H 'Content-Type: application/json' -d '{"offer":"http://localhost:8087/oidc4vci/offer/<id>"}'`
- Open `http://localhost:8087/proof.html` and paste request ID from app
- Return to the app
