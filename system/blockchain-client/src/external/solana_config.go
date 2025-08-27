package external

import (
	"pkg-common/logger"
	"sync"

	"github.com/gagliardetto/solana-go"
)

type Keys struct {
	ContractPublicKey solana.PublicKey
	AccountPublicKey  solana.PublicKey
	AccountPrivateKey solana.PrivateKey
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
		AccountPrivateKey: accountPrivateKey,
	}

	logger.Default().Debugf("Using following public key program id: %s", solanaConfig.ContractPublicKey.String())
	logger.Default().Debugf("Using following public key for signer: %s", solanaConfig.AccountPublicKey.String())

	return &SharedSolanaConfig{
		Mu:   sync.Mutex{},
		Keys: solanaConfig,
	}, nil
}
