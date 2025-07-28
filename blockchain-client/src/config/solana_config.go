package config

import (
	"sync"

	"github.com/gagliardetto/solana-go"
)

type Keys struct {
	ContractPublicKey solana.PublicKey
	AccountPublicKey  solana.PublicKey
}

type SharedSolanaConfig struct {
	// TODO: for zookeeper modifications
	Mu   sync.Mutex
	Keys *Keys
}

func LoadSolanaKeys() (*SharedSolanaConfig, error) {
	contractPrivateKey, err := solana.PrivateKeyFromSolanaKeygenFile("identity_app-keypair.json")
	if err != nil {
		return nil, err
	}

	accountPrivateKey, err := solana.PrivateKeyFromSolanaKeygenFile("id.json")
	if err != nil {
		return nil, err
	}

	solanaConfig := &Keys{
		ContractPublicKey: contractPrivateKey.PublicKey(),
		AccountPublicKey:  accountPrivateKey.PublicKey(),
	}

	return &SharedSolanaConfig{
		Mu:   sync.Mutex{},
		Keys: solanaConfig,
	}, nil
}
