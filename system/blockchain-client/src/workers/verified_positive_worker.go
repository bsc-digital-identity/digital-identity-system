package workers

import (
	"blockchain-client/src/external"
	"blockchain-client/src/types/domain"
	"blockchain-client/src/types/incoming"
	"fmt"

	"blockchain-client/src/zkp"
	"context"
	"encoding/json"
	dtocommon "pkg-common/dto_common"
	"pkg-common/logger"
	"pkg-common/rabbitmq"
	reasoncodes "pkg-common/reason_codes"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/rpc"
	amqp "github.com/rabbitmq/amqp091-go"
)

type VerifiedPositiveWorker struct {
	Config    *external.SharedSolanaConfig
	RpcClient *rpc.Client
	Consumer  rabbitmq.IRabbitmqConsumer
}

func NewVerifiedPositiveWorker() *VerifiedPositiveWorker {
	solanaConfig, err := external.LoadSolanaKeys()
	if err != nil {
		logger.Default().Panicf(err, "Error when loading keys from solana: ")
	}

	return &VerifiedPositiveWorker{
		RpcClient: rpc.New("http://host.docker.internal:8899"),
		Consumer:  rabbitmq.GetConsumer(solanaClientServiceName),
		Config:    solanaConfig,
	}
}

func (sc *VerifiedPositiveWorker) GetServiceName() string {
	return solanaClientServiceName
}

func (sc *VerifiedPositiveWorker) StartService() {
	solanaLogger := logger.Default()
	failurePublisher := rabbitmq.GetPublisher(failureQueuePublisherAlias)
	resultPublisher := rabbitmq.GetPublisher(resultQueuePublisherAlias)

	sc.Consumer.StartConsuming(func(d amqp.Delivery) {
		var message incoming.ZkpVerifiedPositiveDto
		responseFactory := dtocommon.NewZkpProofFailureFactory("", d.Body)

		if err := json.Unmarshal(d.Body, &message); err != nil {
			result := responseFactory.CreateErrorDto(err, reasoncodes.ErrUnmarshal)

			_ = failurePublisher.Publish(result)
			return
		}
		responseFactory = dtocommon.NewZkpProofFailureFactory(message.EventId, d.Body)

		circuitBase := domain.ZkpCircuitBase{}
		zkpResult, err := zkp.CreateZKP(circuitBase)
		if err != nil {
			solanaLogger.Errorf(err, "Failed to create ZKP with user provided data: %d", 10)
			response := responseFactory.CreateErrorDto(err, reasoncodes.ErrProofGeneration)

			_ = failurePublisher.Publish(response)
			return
		}

		signatureChan := make(chan domain.ZkpStorageData)
		errChan := make(chan error)

		go sc.publishZkpToSolana(*zkpResult, errChan, signatureChan)

		var proofReference domain.ZkpStorageData
		select {
		case proofReference = <-signatureChan:
			solanaLogger.Infof("Saved zkp to blockchain with signature: %s", proofReference.Signature.String())
		case err := <-errChan:
			solanaLogger.Errorf(err, "Unable to save the ZKP to the blockchain")

			response := responseFactory.CreateErrorDto(err, reasoncodes.ErrSolana)
			_ = failurePublisher.Publish(response)
			return
		}

		result := dtocommon.ZkpProofResultDto{
			EventId:   message.EventId,
			Signature: proofReference.Signature.String(),
			AccountId: proofReference.Account.String(),
		}

		_ = resultPublisher.Publish(result)
		solanaLogger.Infof("Processed ZKP Verification for %s. Signature: %s, Account: %s", result.EventId, result.Signature, result.AccountId)
	})
}

// TODO: add option for the users to be payers instead of owners
func (sc *VerifiedPositiveWorker) publishZkpToSolana(
	zkpResult zkp.ZkpResult,
	errCh chan error,
	sigCh chan domain.ZkpStorageData) {
	zkpData, err := zkpResult.SerializeBorsh()
	if err != nil {
		errCh <- err
		return
	}

	solanaLogger := logger.Default()
	solanaLogger.Infof("Serialized ZKP data size: %d bytes", len(zkpData))

	sc.createAndPopulateZkpAccount(zkpData, errCh, sigCh)
}

// creates new account and stores zkp data for future retrival
func (sc *VerifiedPositiveWorker) createAndPopulateZkpAccount(
	zkpData []byte,
	errCh chan error,
	sigCh chan domain.ZkpStorageData) {

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
		sc.Config.Keys.ContractPublicKey, // owner = ProgramID
		sc.Config.Keys.AccountPublicKey,  // payer (FROM)
		newAccount.PublicKey(),           // new account (NEW)
	).Build()

	accounts := []*solana.AccountMeta{
		solana.NewAccountMeta(newAccount.PublicKey(), false, true),
		solana.NewAccountMeta(sc.Config.Keys.AccountPublicKey, true, true),
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

	// DEBUG
	{
		msg := tx.Message
		solanaLogger.Infof("AccountKeys (n=%d):", len(msg.AccountKeys))
		for i, k := range msg.AccountKeys {
			solanaLogger.Infof("  [%d] %s", i, k.String())
		}
		for i, ix := range msg.Instructions {
			solanaLogger.Infof("IX #%d: programIDIndex=%d accounts=%v dataLen=%d", i, ix.ProgramIDIndex, ix.Accounts, len(ix.Data))
			// sanity check:
			for _, idx := range ix.Accounts {
				if int(idx) >= len(msg.AccountKeys) {
					solanaLogger.Errorf(nil, "BAD INDEX: ix#%d uses account index %d (n=%d)", i, idx, len(msg.AccountKeys))
				}
			}
			if int(ix.ProgramIDIndex) >= len(msg.AccountKeys) {
				solanaLogger.Errorf(nil, "BAD PROGRAM INDEX: ix#%d programIDIndex=%d (n=%d)", i, ix.ProgramIDIndex, len(msg.AccountKeys))
			}
		}
	}

	sim, simErr := sc.RpcClient.SimulateTransaction(context.Background(), tx)
	if simErr != nil {
		errCh <- fmt.Errorf("simulate call: %w", simErr)
		return
	}
	if sim.Value.Err != nil {
		for _, l := range sim.Value.Logs {
			logger.Default().Debug(l)
		}
		errCh <- fmt.Errorf("simulate err: %+v", sim.Value.Err)
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

	sigCh <- domain.ZkpStorageData{
		Signature: transactionSignature,
		Account:   newAccount.PublicKey(),
	}
}

func (sc *VerifiedPositiveWorker) createZkpAccount(
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
		sc.Config.Keys.ContractPublicKey, // owner = ProgramID
		sc.Config.Keys.AccountPublicKey,  // payer (FROM)
		newAccount.PublicKey(),           // new account (NEW)
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
