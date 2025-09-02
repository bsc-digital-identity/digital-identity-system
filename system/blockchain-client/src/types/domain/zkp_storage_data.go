package domain

import "github.com/gagliardetto/solana-go"

type ZkpStorageData struct {
	Account   solana.PublicKey
	Signature solana.Signature
}
