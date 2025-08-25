#!/bin/sh

PROJECT_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/../../" && pwd -P)
BLOCKCHAIN_CLIENT_SRC_DIR="${PROJECT_ROOT}/system/blockchain-client/"
API_SRC_DIR="${PROJECT_ROOT}/system/api/"

export PATH=$PATH:$(go env GOPATH)/bin

if ! swag --version &> /dev/null
then
    echo "Swag CLI not found, installing..."
    go install github.com/swaggo/swag/cmd/swag@latest

    if ! swag --version &> /dev/null
    then
        echo "Failed to install swag CLI. Please ensure GOPATH/bin is in your PATH."
        exit 1
    fi
fi

swag --version

echo "Generating Swagger docs..."

cd $BLOCKCHAIN_CLIENT_SRC_DIR
swag init -g src/main.go -o src/docs

cd $API_SRC_DIR 
swag init -g src/main.go -o src/docs

echo "Swagger docs generated succesfully"