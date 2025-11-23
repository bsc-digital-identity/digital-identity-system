#!/usr/bin/env bash
set -euo pipefail

# ---------- helpers ----------
log() { printf 'MESSAGE: %s\n' "$*"; }

detect_lan_host_ip() {
  if [[ -n "${LAN_HOST_IP:-}" ]]; then
    echo "${LAN_HOST_IP}"
    return 0
  fi

  # Prefer IP from default route lookup
  if command -v ip >/dev/null 2>&1; then
    if route_ip=$(ip route get 1.1.1.1 2>/dev/null | awk '{for(i=1;i<=NF;i++) if ($i=="src") {print $(i+1); exit}}'); [[ -n "${route_ip}" ]]; then
      echo "${route_ip}"
      return 0
    fi
  fi

  # Fallback to first non-loopback address from hostname -I
  if command -v hostname >/dev/null 2>&1; then
    if host_ip=$(hostname -I 2>/dev/null | awk '{print $1}'); [[ -n "${host_ip}" && "${host_ip}" != "127.0.0.1" ]]; then
      echo "${host_ip}"
      return 0
    fi
  fi

  echo "127.0.0.1"
}

LAN_HOST_IP_VALUE="$(detect_lan_host_ip)"
log "Using LAN_HOST_IP=${LAN_HOST_IP_VALUE}"

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

# Trap on Ctrl+C or termination (not ERR, so we see the real error)
trap cleanup INT TERM

# ---------- preflight ----------
command -v solana >/dev/null || { echo "ERROR: 'solana' CLI not found in PATH." >&2; exit 1; }

# Use local RPC for all solana CLI calls
SOLANA_URL="http://${LAN_HOST_IP_VALUE}:8899"

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

# Extra guard: ensure the validator didn't crash right after RPC probe
if ! ps -p "${VALIDATOR_PID}" >/dev/null 2>&1; then
  echo "ERROR: solana-test-validator is not running (likely Windows symlink privilege). Enable Developer Mode and re-run." >&2
  exit 1
fi

# ---------- build & deploy smart contract ----------
chmod +x dev_tools/scripts/smart_contract.sh

log "Running smart_contract.sh check"
dev_tools/scripts/smart_contract.sh check

log "Deploying program..."
# Parse both "Keypair Path" and "Program Id" robustly, preserving Windows paths with ':'.
# We write to a temporary .env then move it atomically when complete.
dev_tools/scripts/smart_contract.sh deploy \
| tee >(awk '
  {
    gsub(/\r/, "", $0)  # strip CR if present
  }
  /Keypair Path:/ {
    line = $0
    sub(/^[^:]+:\s*/, "", line)   # drop "Keypair Path: "
    print "KEYPAIR_PATH=" line
  }
  /Program Id:/ {
    line = $0
    sub(/^[^:]+:\s*/, "", line)   # drop "Program Id: "
    print "PROGRAM_ID=" line
  }
' > "${ENV_TMP}")

# Ensure we captured both values before finalizing .env
if ! (grep -q '^KEYPAIR_PATH=' "${ENV_TMP}" && grep -q '^PROGRAM_ID=' "${ENV_TMP}"); then
  echo "ERROR: Failed to parse KEYPAIR_PATH and PROGRAM_ID from deploy output." >&2
  cat "${ENV_TMP}" >&2 || true
  exit 1
fi
mv -f "${ENV_TMP}" .env
printf 'LAN_HOST_IP=%s\nSOLANA_URL=http://%s:8899\n' "${LAN_HOST_IP_VALUE}" "${LAN_HOST_IP_VALUE}" >> .env
log "Wrote .env with KEYPAIR_PATH, PROGRAM_ID, LAN_HOST_IP=${LAN_HOST_IP_VALUE}, SOLANA_URL=http://${LAN_HOST_IP_VALUE}:8899"

# ---------- fund local wallet ----------
mkdir -p ./system/blockchain-client
if [[ ! -f "${HOME}/.config/solana/id.json" ]]; then
  echo "ERROR: ${HOME}/.config/solana/id.json not found. Did you run 'solana-keygen new'?" >&2
  exit 1
fi

cp -f "${HOME}/.config/solana/id.json" ./system/blockchain-client/id.json

PUBKEY=$(solana-keygen pubkey ./system/blockchain-client/id.json)
log "Requesting airdrop for ${PUBKEY}..."

max_try=10
for try in $(seq 1 $max_try); do
  if solana --url "${SOLANA_URL}" airdrop 2 "${PUBKEY}"; then
    log "Airdrop complete."
    break
  fi
  if [[ $try -eq $max_try ]]; then
    echo "ERROR: airdrop failed after ${max_try} attempts." >&2
    exit 1
  fi
  sleep 1
done

# ---------- docker compose ----------
log "setup ready, starting docker"
# Run in the foreground so Ctrl+C triggers our trap, then cleanup runs.
"${DCMD[@]}" up --build

# If compose exits normally (containers stop), clean up the validator as well.
cleanup
