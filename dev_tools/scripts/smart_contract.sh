#!/bin/bash

# resolve directories
PROJECT_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../" && pwd -P)

SMART_CONTRACT_DIR="${PROJECT_ROOT}/system/smart-contract"
BLOCKCHAIN_CLIENT_DIR="${PROJECT_ROOT}/system/blockchain-client"
CLIENT_DIR="${SMART_CONTRACT_DIR}/client"
PROGRAM_DIR="${SMART_CONTRACT_DIR}/program"
DIST_DIR="${SMART_CONTRACT_DIR}/dist"

build_sbf() {
    cargo build-sbf --manifest-path="${PROGRAM_DIR}/Cargo.toml" --sbf-out-dir="${DIST_DIR}/program"
}

check_toolchain() {
    echo "You are running with following toolchain, please install missing tools:"
    cargo --version
    rustc --version
    solana-test-validator --version
    solana --version
    echo "All is good!!!"
}

wait_for_validator() {
    echo "Waiting for validator to be ready..."
    until curl --output /dev/null --silent --head --fail http://127.0.0.1:8899; do
        printf '.'
        sleep 1
    done
    echo "Validator is ready!"
}

cleanup() {
    for pid in "${pids[@]}"; do
        echo "Stopping process on pid $pid"
        kill $pid
    done
    echo "All background processes stopped"
    exit 0
}

pids=()

case $1 in
    "check")
        check_toolchain
        ;;
    "build"|"build-sbf")
        build_sbf
        ;;
    "deploy")
        echo "solana-test-validator must be started before running this command or it will fail"
        build_sbf
        solana config set --url http://127.0.0.1:8899
        solana program deploy "${DIST_DIR}/program/identity_app.so" --program-id "${DIST_DIR}/program/identity_app-keypair.json"
        cp "${DIST_DIR}/program/identity_app-keypair.json" "${BLOCKCHAIN_CLIENT_DIR}/"
        ;;
    "client")
        (cd "${CLIENT_DIR}" && cargo run "${DIST_DIR}/program/identity_app-keypair.json")
        ;;
    "clean")
        (cd "${PROGRAM_DIR}" && cargo clean)
        (cd "${CLIENT_DIR}" && cargo clean)
        rm -rf "${DIST_DIR}"
        ;;
    *)
        echo "usage: $0 [check|build|deploy|client|clean]"
        ;;
esac