#!/bin/bash

# resolve directories
PROJECT_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../" && pwd -P)

TEST_DIR_NAME="test"
BLOCKCHAIN_CLIENT_TEST_DIR="${PROJECT_ROOT}/system/blockchain-client/${TEST_DIR_NAME}"
API_TEST_DIR="${PROJECT_ROOT}/system/api/${TEST_DIR_NAME}"
PKG_TEST_DIR="${PROJECT_ROOT}/system/pkg/${TEST_DIR_NAME}"

run_all() {
    cd $BLOCKCHAIN_CLIENT_TEST_DIR/..
    go clean -cache && go clean -testcache
    go test -v $BLOCKCHAIN_CLIENT_TEST_DIR
    
    cd $API_TEST_DIR/..
    go clean -cache && go clean -testcache
    go test -v $API_TEST_DIR
    
    cd $PKG_TEST_DIR/..
    go clean -cache && go clean -testcache
    go test -v $PKG_TEST_DIR
}

becnmark_all() {
    cd $BLOCKCHAIN_CLIENT_TEST_DIR/..
    go clean -cache && go clean -testcache
    go test -bench . -run notest $BLOCKCHAIN_CLIENT_TEST_DIR
    
    cd $API_TEST_DIR/..
    go clean -cache && go clean -testcache
    go test -bench . -run notest $API_TEST_DIR
    
    cd $PKG_TEST_DIR/..
    go clean -cache && go clean -testcache
    go test -bench . -run notest $PKG_TEST_DIR
}

case $1 in
    "test")
        run_all
        ;;
    "bench")
        becnmark_all
    ;;
    *)
        echo "usage: $0 [test|bench]"
        ;;
esac