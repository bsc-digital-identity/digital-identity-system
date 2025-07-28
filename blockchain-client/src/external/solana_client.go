package external

import (
	"blockchain-client/src/config"
	"blockchain-client/src/zkp"
	"context"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type SolanaClient struct {
	Config    *config.SharedSolanaConfig
	RpcClient *rpc.Client
}

// add option for the users to be payers instead of owners
func (sc *SolanaClient) PublishZkpToSolana(
	zkpResult zkp.ZkpResult,
	errCh chan error,
	sigCh chan solana.Signature) {
	zkpData, err := zkpResult.SerializeBorsh()
	if err != nil {
		errCh <- err
		return
	}

	// lock mutex to read correct values at current time
	sc.Config.Mu.Lock()

	accounts := []*solana.AccountMeta{
		solana.NewAccountMeta(sc.Config.Keys.AccountPublicKey, true, true),
		solana.NewAccountMeta(sc.Config.Keys.ContractPublicKey, false, false),
	}

	instruction := &solana.GenericInstruction{
		ProgID:        sc.Config.Keys.ContractPublicKey,
		AccountValues: accounts,
		DataBytes:     zkpData,
	}
	sc.Config.Mu.Unlock()

	recent, err := sc.RpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentConfirmed)
	if err != nil {
		errCh <- err
		return
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{instruction},
		recent.Value.Blockhash,
		// Here replace with actual payer
		solana.TransactionPayer(sc.Config.Keys.AccountPublicKey),
	)
	if err != nil {
		errCh <- err
		return
	}

	transactionSignature, err := sc.RpcClient.SendTransactionWithOpts(
		context.Background(),
		tx,
		rpc.TransactionOpts{
			SkipPreflight:       false,
			PreflightCommitment: rpc.CommitmentFinalized,
		},
	)
	if err != nil {
		errCh <- err
		return
	}

	sigCh <- transactionSignature
}
