package external

import (
	"blockchain-client/src/config"
	"blockchain-client/src/zkp"
	"bytes"

	"github.com/gagliardetto/solana-go/rpc"
)

type SolanaClient struct {
	Config    *config.SharedSolanaConfig
	RpcClient *rpc.Client
}

// add option for the users to be payers instead of owners
func (sc *SolanaClient) PublishZkpToSolana(zkpResult zkp.ZkpResult) error {
	var instructionData bytes.Buffer

	err :=
}

