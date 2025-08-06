package external

import (
	"blockchain-client/src/config"
	"blockchain-client/src/queues"

	"blockchain-client/src/zkp"
	"context"
	"encoding/json"
	"pkg-common/logger"
	"pkg-common/rabbitmq"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SolanaClient struct {
	Config    *config.SharedSolanaConfig
	RpcClient *rpc.Client
	Consumer  rabbitmq.IRabbitmqConsumer
}

func NewSolanaClient(config *config.SharedSolanaConfig) *SolanaClient {
	return &SolanaClient{
		RpcClient: rpc.New("http://host.docker.internal:8899"),
		Consumer:  rabbitmq.GetConsumer("VerifiedPositiveConsumer"),
		Config:    config,
	}

}

func (sc *SolanaClient) StartService() {
	solanaLogger := logger.Default()
	failurePublisher := rabbitmq.GetPublisher(rabbitmq.PublisherAlias("IdentityFailurePublisher"))
	resultPublisher := rabbitmq.GetPublisher(rabbitmq.PublisherAlias("IdentityResultsPublisher"))

	sc.Consumer.StartConsuming(func(d amqp.Delivery) {
		var req queues.ZeroKnowledgeProofVerificationRequest
		err := json.Unmarshal(d.Body, &req)
		if err != nil {
			result := queues.ZeroKnowledgeProofVerificationResponse{
				IdentityId:   req.IdentityId,
				IsProofValid: false,
				Error:        "unmarshal: " + err.Error(),
			}

			_ = failurePublisher.Publish(result)
		}

		// TODO: mock data read from reqesut
		// replace to read from request
		zkpResult, err := zkp.CreateZKP(10, 10, 1990)
		if err != nil {
			solanaLogger.Errorf(err, "Failed to create ZKP with user provided data: %d", 10)
			return
		}

		signatureChan := make(chan solana.Signature)
		errChan := make(chan error)

		go sc.PublishZkpToSolana(*zkpResult, errChan, signatureChan)

		var signature solana.Signature
		select {
		case signature = <-signatureChan:
			solanaLogger.Infof("Saved zkp to blockchain with signature: %s", signature.String())
		case err := <-errChan:
			solanaLogger.Errorf(err, "Unable to save the ZKP to the blockchain")
		}

		result := queues.MockZKPVerification(req, signature)

		_ = resultPublisher.Publish(req)
		solanaLogger.Infof("Processed ZKP Verification for %s. ProofReference: %s", req.IdentityId, result.ProofReference)
	})
}

// TODO: add option for the users to be payers instead of owners
func (sc *SolanaClient) PublishZkpToSolana(
	zkpResult zkp.ZkpResult,
	errCh chan error,
	sigCh chan solana.Signature) {
	zkpData, err := zkpResult.SerializeBorsh()
	if err != nil {
		errCh <- err
		return
	}

	solanaLogger := logger.Default()
	solanaLogger.Infof("Serialized ZKP data size: %d bytes", len(zkpData))

	sc.CreateAndPopulateZkpAccount(zkpData, errCh, sigCh)
}

// creates new account and stores zkp data for future retrival
func (sc *SolanaClient) CreateAndPopulateZkpAccount(
	zkpData []byte,
	errCh chan error,
	sigCh chan solana.Signature) {

	solanaLogger := logger.Default()
	space := calculateRequiredAccountSpace(zkpData)
	solanaLogger.Infof("ZKP data size: %d bytes, allocated space: %d bytes", len(zkpData), space)

	rent, err := sc.RpcClient.GetMinimumBalanceForRentExemption(
		context.Background(),
		space,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		errCh <- err
		return
	}
	solanaLogger.Infof("Required rent for account: %d lamports", rent)

	newAccount, err := solana.NewRandomPrivateKey()
	if err != nil {
		errCh <- err
		return
	}
	solanaLogger.Infof("Generated new account: %s", newAccount.PublicKey().String())

	// mutex lock to read correct values at current time
	sc.Config.Mu.Lock()

	createAccountInstruction := system.NewCreateAccountInstruction(
		rent,
		space,
		sc.Config.Keys.ContractPublicKey,
		sc.Config.Keys.AccountPublicKey,
		newAccount.PublicKey(),
	).Build()

	accounts := []*solana.AccountMeta{
		solana.NewAccountMeta(newAccount.PublicKey(), true, true),
	}

	zkpInstruction := solana.NewInstruction(
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
		[]solana.Instruction{createAccountInstruction, zkpInstruction},
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

	transactionSignature, err := sc.RpcClient.SendTransactionWithOpts(
		context.Background(),
		tx,
		rpc.TransactionOpts{
			SkipPreflight:       false,
			PreflightCommitment: rpc.CommitmentFinalized,
		},
	)
	if err != nil {
		solanaLogger.Errorf(err, "Failed to send combined transaction")
		solanaLogger.Debugf("ZKP data size: %d bytes, allocated space: %d bytes", len(zkpData), space)
		errCh <- err
		return
	}

	solanaLogger.Infof("Successfully sent combined transaction: %s", transactionSignature)
	sigCh <- transactionSignature
}

func (sc *SolanaClient) CreateZkpAccount(
	zkpData []byte,
	errCh chan error,
	accountCh chan solana.PrivateKey) {
	solanaLogger := logger.Default()
	space := calculateRequiredAccountSpace(zkpData)

	solanaLogger.Infof("ZKP data size: %d bytes, allocated space: %d bytes", len(zkpData), space)

	rent, err := sc.RpcClient.GetMinimumBalanceForRentExemption(
		context.Background(),
		space,
		rpc.CommitmentFinalized,
	)
	if err != nil {
		errCh <- err
		return
	}

	solanaLogger.Infof("[INFO]: Required rent for account: %d lamports", rent)

	newAccount, err := solana.NewRandomPrivateKey()
	if err != nil {
		errCh <- err
		return
	}

	createAccountInstruction := system.NewCreateAccountInstruction(
		rent,
		space,
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
		solanaLogger.Errorf(err, "Failed to create account")
		solanaLogger.Debugf("Requested space: %d bytes, rent: %d lamports", space, rent)
		errCh <- err
		return
	}

	solanaLogger.Infof("Successfully created account with %d bytes of space", space)

	accountCh <- newAccount
}

func calculateRequiredAccountSpace(data []byte) uint64 {
	// calculate space to store whole ZKP data
	// with enough buffer for it to fit with metadata
	dataSize := len(data)

	var totalSize int
	if dataSize > 10000 {
		totalSize = int(float64(dataSize) * 1.5)
	} else if dataSize > 1000 {
		totalSize = dataSize + 2048
	} else {
		totalSize = dataSize + 1024
	}

	// round to 8 bytes
	if totalSize%8 != 0 {
		totalSize += 8 - (totalSize % 8)
	}

	// minimum is 2048
	if totalSize < 2048 {
		totalSize = 2048
	}

	logger.Default().Debugf("Data size: %d, calculated space: %d", dataSize, totalSize)
	return uint64(totalSize)
}
