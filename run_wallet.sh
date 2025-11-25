#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
WALLET_DIR="${REPO_ROOT}/system/zk-wallet-go"
ENV_FILE="${WALLET_DIR}/.env"
DEFAULT_ENV="${WALLET_DIR}/default.env"

if [[ ! -f "${ENV_FILE}" ]]; then
  echo "Missing ${ENV_FILE}. Copy ${DEFAULT_ENV} and fill in your DSNET/OIDC credentials." >&2
  exit 1
fi

cd "${WALLET_DIR}"
echo "Starting zk-wallet-go from ${WALLET_DIR}"
exec go run cmd/wallet-server/main.go
