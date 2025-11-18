package external

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"pkg-common/logger"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
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
	// 1) Program ID
	programIDStr := os.Getenv("PROGRAM_ID")
	if programIDStr == "" {
		return nil, fmt.Errorf("PROGRAM_ID env var is not set (use your deployed program id, e.g. HxSN1y...)")
	}
	programID, err := solana.PublicKeyFromBase58(programIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid PROGRAM_ID %q: %w", programIDStr, err)
	}

	// 2) Payer keypair
	keypairPath := os.Getenv("PAYER_KEYPAIR_PATH")
	if keypairPath == "" {
		homeDir, _ := os.UserHomeDir()
		keypairPath = filepath.Join(homeDir, ".zkpconfig", "solana", "id.json")
	}
	payerPriv, err := solana.PrivateKeyFromSolanaKeygenFile(keypairPath)
	if err != nil {
		return nil, fmt.Errorf("reading payer keypair from %s failed: %w", keypairPath, err)
	}

	cfg := &Keys{
		ContractPublicKey: programID,
		AccountPublicKey:  payerPriv.PublicKey(),
		AccountPrivateKey: payerPriv,
	}

	logger.Default().Debugf("ProgramID (ContractPublicKey): %s", cfg.ContractPublicKey.String())
	logger.Default().Debugf("Payer (AccountPublicKey): %s", cfg.AccountPublicKey.String())

	return &SharedSolanaConfig{
		Mu:   sync.Mutex{},
		Keys: cfg,
	}, nil
}

func (sc *SharedSolanaConfig) ValidateProgramExecutable(ctx context.Context, rpcClient *rpc.Client) error {
	acc, err := rpcClient.GetAccountInfo(ctx, sc.Keys.ContractPublicKey)
	if err != nil {
		return fmt.Errorf("GetAccountInfo(program) failed: %w", err)
	}
	if acc == nil || acc.Value == nil || !acc.Value.Executable {
		return fmt.Errorf("ContractPublicKey %s is not an executable account (this is NOT a program id)", sc.Keys.ContractPublicKey)
	}
	return nil
}
