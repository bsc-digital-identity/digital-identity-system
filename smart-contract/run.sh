#!/bin/bash

build_sbf() {
    cargo build-sbf --manifest-path=program/Cargo.toml --sbf-out-dir=dist/program
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
    echo "All background processess stopped"
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
        solana program deploy dist/program/identity_app.so --program-id dist/program/identity_app-keypair.json
        cp dist/program/identity_app-keypair.json ../blockchain-client/
        ;;
    "client")
        (cd client/; cargo run /dist/program/helloworld-keypair.json)
        ;;
    "clean")
        (cd program/; cargo clean)
        (cd client/; cargo clean)
        rm -rf dist/
        ;;
    *)
        echo "usage: $0 [check|build|deploy|client|clean]"
        ;;
esac