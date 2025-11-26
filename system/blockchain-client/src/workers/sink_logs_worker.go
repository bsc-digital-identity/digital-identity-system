package workers

import (
	"blockchain-client/src/external"
	"context"
	"encoding/json"
	"fmt"

	"pkg-common/logger"
	"pkg-common/rabbitmq"
	logger_message "pkg-common/utilities/logger"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	amqp "github.com/rabbitmq/amqp091-go"
)

type BlockchainClientLogSink struct {
	Config    *external.SharedSolanaConfig
	RpcClient *rpc.Client
	Consumer  rabbitmq.IRabbitmqConsumer
	logger    *logger.Logger
}

func NewBlockchainClientLogSink() *BlockchainClientLogSink {
	solanaConfig, err := external.LoadSolanaKeys()
	if err != nil {
		panic(fmt.Sprintf("Error when loading keys from solana: %v", err))
	}

	return &BlockchainClientLogSink{
		RpcClient: rpc.New("http://host.docker.internal:8899"),
		Consumer:  rabbitmq.GetConsumer(logConsumerAlias),
		Config:    solanaConfig,
		logger:    logger.New(),
	}
}

func (lw *BlockchainClientLogSink) GetServiceName() string {
	return logConsumerAlias
}

func (lw *BlockchainClientLogSink) StartService() {

	lw.Consumer.StartConsuming(func(d amqp.Delivery) {
		var logMessage logger_message.LoggerMessage

		if err := json.Unmarshal(d.Body, &logMessage); err != nil {
			return
		}

		if err := lw.storeLogToSolana(logMessage); err != nil {
			return
		}
	})
}

func (lw *BlockchainClientLogSink) storeLogToSolana(logMessage logger_message.LoggerMessage) error {
	logData, err := logMessage.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize log message: %w", err)
	}

	instructionData := append([]byte("LOG:"), logData...)

	lw.Config.Mu.Lock()
	defer lw.Config.Mu.Unlock()

	accounts := []*solana.AccountMeta{
		solana.NewAccountMeta(lw.Config.Keys.AccountPublicKey, true, true),
	}

	logInstruction := solana.NewInstruction(
		lw.Config.Keys.ContractPublicKey,
		accounts,
		instructionData,
	)

	latest, err := lw.RpcClient.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return fmt.Errorf("failed to get latest blockhash: %w", err)
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{logInstruction},
		latest.Value.Blockhash,
		solana.TransactionPayer(lw.Config.Keys.AccountPublicKey),
	)
	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	_, err = tx.Sign(func(pk solana.PublicKey) *solana.PrivateKey {
		if pk.Equals(lw.Config.Keys.AccountPublicKey) {
			return &lw.Config.Keys.AccountPrivateKey
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to sign transaction: %w", err)
	}

	sim, simErr := lw.RpcClient.SimulateTransaction(context.Background(), tx)
	if simErr != nil {
		return fmt.Errorf("simulate call failed: %w", simErr)
	}
	if sim.Value.Err != nil {
		for _, l := range sim.Value.Logs {
			lw.logger.Debug(l)
		}
		return fmt.Errorf("simulation failed: %+v", sim.Value.Err)
	}

	transactionSignature, err := lw.RpcClient.SendTransactionWithOpts(
		context.Background(),
		tx,
		rpc.TransactionOpts{
			SkipPreflight:       false,
			PreflightCommitment: rpc.CommitmentFinalized,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send transaction: %w", err)
	}

	fmt.Sprintf("Log stored to blockchain with signature: %s", transactionSignature)
	return nil
}
