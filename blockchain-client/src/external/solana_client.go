package external

import (
	"blockchain-client/src/config"
	"blockchain-client/src/zkp"
	"context"
	"log"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
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

	newAccountChannel := make(chan solana.PrivateKey)
	errorChannel := make(chan error)

	go sc.CreateZkpAccount(zkpData, errorChannel, newAccountChannel)

	var newAccount solana.PrivateKey
	select {
	case newAccount = <-newAccountChannel:
		log.Printf("[INFO]: Created new account: %s", newAccount.PublicKey().String())
	case err := <-errorChannel:
		errCh <- err
		return
	}

	// lock mutex to read correct values at current time
	sc.Config.Mu.Lock()
	accounts := []*solana.AccountMeta{
		solana.NewAccountMeta(newAccount.PublicKey(), true, true),
	}

	instruction := solana.NewInstruction(
		sc.Config.Keys.ContractPublicKey,
		accounts,
		zkpData,
	)

	sc.Config.Mu.Unlock()

	latest, err := sc.RpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		errCh <- err
		return
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{instruction},
		latest.Value.Blockhash,
		// Here replace with actual payer
		solana.TransactionPayer(sc.Config.Keys.AccountPublicKey),
	)
	if err != nil {
		errCh <- err
		return
	}

	_, err = tx.Sign(func(pk solana.PublicKey) *solana.PrivateKey {
		if pk.Equals(sc.Config.Keys.AccountPublicKey) {
			return &sc.Config.Keys.AccountPrivateKey
		}
		if pk.Equals(newAccount.PublicKey()) {
			return &newAccount
		}
		return nil
	})

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

func (sc *SolanaClient) CreateZkpAccount(
	zkpData []byte,
	errCh chan error,
	accountCh chan solana.PrivateKey) {
	space := calculateRequiredAccountSpace(zkpData)
	rent, err := sc.RpcClient.GetMinimumBalanceForRentExemption(
		context.Background(),
		space,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		errCh <- err
		return
	}

	newAccount, err := solana.NewRandomPrivateKey()
	if err != nil {
		errCh <- err
		return
	}

	createAccountInstruction := system.NewCreateAccountInstruction(
		rent,
		uint64(space),
		sc.Config.Keys.ContractPublicKey,
		sc.Config.Keys.AccountPublicKey,
		newAccount.PublicKey(),
	).Build()

	latest, err := sc.RpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		errCh <- err
		return
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{createAccountInstruction},
		latest.Value.Blockhash,
		solana.TransactionPayer(sc.Config.Keys.AccountPublicKey),
	)
	if err != nil {
		errCh <- err
		return
	}

	_, err = tx.Sign(func(pk solana.PublicKey) *solana.PrivateKey {
		if pk.Equals(sc.Config.Keys.AccountPublicKey) {
			return &sc.Config.Keys.AccountPrivateKey
		}
		if pk.Equals(newAccount.PublicKey()) {
			return &newAccount
		}
		return nil
	})
	if err != nil {
		errCh <- err
		return
	}

	_, err = sc.RpcClient.SendTransactionWithOpts(
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

	accountCh <- newAccount
}

func calculateRequiredAccountSpace(data []byte) uint64 {
	totalSize := len(data) + 8

	if totalSize%8 != 0 {
		totalSize += 8 - (totalSize % 8)
	}

	return uint64(totalSize)
}
