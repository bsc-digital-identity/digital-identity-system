#!/usr/bin/env bash
set -euo pipefail

# ---------- helpers ----------
log() { printf 'MESSAGE: %s\n' "$*"; }

# Prefer v2 `docker compose` if available, else fall back to v1 `docker-compose`
if command -v docker &>/dev/null && docker compose version &>/dev/null; then
  DCMD=(docker compose)
elif command -v docker-compose &>/dev/null; then
  DCMD=(docker-compose)
else
  echo "ERROR: Neither 'docker compose' nor 'docker-compose' is available." >&2
  exit 1
fi

VALIDATOR_PID=""
ENV_TMP=".env.tmp"

cleanup() {
  # Don't let cleanup fail the trap
  set +e
  log "Stopping Solana test validator and Docker containers..."

  if [[ -n "${VALIDATOR_PID}" ]] && ps -p "${VALIDATOR_PID}" &>/dev/null; then
    kill "${VALIDATOR_PID}"
    # Give it a moment to exit gracefully
    wait "${VALIDATOR_PID}" 2>/dev/null
    log "solana-test-validator (pid ${VALIDATOR_PID}) stopped"
  else
    # Fallback if we somehow lost the PID
    pkill -f solana-test-validator 2>/dev/null || true
    log "solana-test-validator stopped (by name)"
  fi

  # Bring containers down (ok if none were up)
  "${DCMD[@]}" down || true
  log "docker compose down"

  # Clean tmp env file if we never finalized it
  [[ -f "${ENV_TMP}" ]] && rm -f "${ENV_TMP}"

  exit 0
}

# Trap on Ctrl+C, termination, or any unhandled error
trap cleanup INT TERM ERR

# ---------- preflight ----------
command -v solana >/dev/null || { echo "ERROR: 'solana' CLI not found in PATH." >&2; exit 1; }

# Use local RPC for all solana CLI calls
SOLANA_URL="http://127.0.0.1:8899"

# ---------- start local validator ----------
log "Starting solana-test-validator..."
solana-test-validator --reset --quiet --rpc-port 8899 --faucet-port 9900 &
VALIDATOR_PID=$!
log "solana-test-validator started (pid ${VALIDATOR_PID})"

# Wait until RPC is responsive (max ~30s)
log "Waiting for RPC at ${SOLANA_URL}..."
for i in {1..60}; do
  if solana --url "${SOLANA_URL}" cluster-version >/dev/null 2>&1; then
    log "RPC is up."
    break
  fi
  sleep 0.5
  if [[ $i -eq 60 ]]; then
    echo "ERROR: solana-test-validator RPC did not become ready in time." >&2
    exit 1
  fi
done

# ---------- build & deploy smart contract ----------
chmod +x dev_tools/scripts/smart_contract.sh

log "Running smart_contract.sh check"
dev_tools/scripts/smart_contract.sh check

log "Deploying program..."
# Parse both "Keypair Path" and "Program Id" robustly, regardless of order/spacing/CRLF.
# We write to a temporary .env then move it atomically when complete.
dev_tools/scripts/smart_contract.sh deploy \
| tee >(awk -F': *' '
    /Keypair Path|Program Id/ {
      # Normalize CR if present
      gsub(/\r/, "")
      if ($1 ~ /Keypair Path/) { print "KEYPAIR_PATH=" $2 }
      else if ($1 ~ /Program Id/) { print "PROGRAM_ID=" $2 }
    }
  ' > "${ENV_TMP}"
)

# Ensure we captured both values before finalizing .env
if ! (grep -q '^KEYPAIR_PATH=' "${ENV_TMP}" && grep -q '^PROGRAM_ID=' "${ENV_TMP}"); then
  echo "ERROR: Failed to parse KEYPAIR_PATH and PROGRAM_ID from deploy output." >&2
  cat "${ENV_TMP}" >&2 || true
  exit 1
fi
mv -f "${ENV_TMP}" .env
log "Wrote .env with KEYPAIR_PATH and PROGRAM_ID"

# ---------- fund local wallet ----------
mkdir -p ./system/blockchain-client
if [[ ! -f "${HOME}/.config/solana/id.json" ]]; then
  echo "ERROR: ${HOME}/.config/solana/id.json not found. Did you run 'solana-keygen new'?" >&2
  exit 1
fi

cp -f "${HOME}/.config/solana/id.json" ./system/blockchain-client/id.json

PUBKEY=$(solana-keygen pubkey ./system/blockchain-client/id.json)
log "Requesting airdrop for ${PUBKEY}..."
solana --url "${SOLANA_URL}" airdrop 2 "${PUBKEY}"
log "Airdrop complete."

# ---------- docker compose ----------
log "setup ready, starting docker"
# Run in the foreground so Ctrl+C triggers our trap, then cleanup runs.
"${DCMD[@]}" up --build

# If compose exits normally (containers stop), clean up the validator as well.
cleanup