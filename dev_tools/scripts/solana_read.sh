#!/bin/bash

case $1 in
    "history")
        solana transaction-history $2 --url http://localhost:8899 --limit 100
        ;;
    "account")
        solana account $2 --url http://localhost:8899 --output json
    ;;
    *)
        echo "usage: $0 [history <ACCOUNT> | account <TRANSACTION_SIGNATURE>]"
        ;;
esac

