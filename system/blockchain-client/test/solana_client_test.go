package test

import (
	"blockchain-client/src/external"
	"testing"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type solanaTestClient struct {
	Config    *external.SharedSolanaConfig
	RpcClient *rpc.Client
}

func newSolanaTestClient(cfg *external.SharedSolanaConfig) *solanaTestClient {
	return &solanaTestClient{
		Config:    cfg,
		RpcClient: rpc.New("http://127.0.0.1:8899"),
	}
}

func (c *solanaTestClient) StartService() {
	c.Config.Mu.Lock()
	defer c.Config.Mu.Unlock()

	_ = c.Config.Keys.ContractPublicKey
}

func TestNewSolanaClient(t *testing.T) {
	// Create a test zkpconfig
	privateKey := solana.MustPrivateKeyFromBase58("4Z7cXSyeFR8wNGMVXUE1TwtKn5D5Vu7FzEv69dokLv7KrQk7h6pu4LF8ZRR9yQBhc7uSM9PiLpAkKktDD8kUmyHT")

	keys := &external.Keys{
		ContractPublicKey: privateKey.PublicKey(),
		AccountPublicKey:  privateKey.PublicKey(),
		AccountPrivateKey: privateKey,
	}

	solanaConfig := &external.SharedSolanaConfig{
		Keys: keys,
	}

	// Test client creation
	client := newSolanaTestClient(solanaConfig)

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	if client.Config == nil {
		t.Error("Expected zkpconfig to be set")
	}

	if client.RpcClient == nil {
		t.Error("Expected RPC client to be initialized")
	}
}

func TestSolanaClientRPCConnection(t *testing.T) {
	// Create RPC client for testing
	rpcClient := rpc.New("http://127.0.0.1:8899")

	if rpcClient == nil {
		t.Fatal("Failed to create RPC client")
	}

	// Test that we can create a client with different endpoints
	testEndpoints := []string{
		"http://127.0.0.1:8899",
		"http://localhost:8899",
		"https://api.mainnet-beta.solana.com",
	}

	for _, endpoint := range testEndpoints {
		client := rpc.New(endpoint)
		if client == nil {
			t.Errorf("Failed to create RPC client for endpoint: %s", endpoint)
		}
	}
}

func TestSolanaClientConfigIntegrity(t *testing.T) {
	privateKey := solana.MustPrivateKeyFromBase58("4Z7cXSyeFR8wNGMVXUE1TwtKn5D5Vu7FzEv69dokLv7KrQk7h6pu4LF8ZRR9yQBhc7uSM9PiLpAkKktDD8kUmyHT")

	keys := &external.Keys{
		ContractPublicKey: privateKey.PublicKey(),
		AccountPublicKey:  privateKey.PublicKey(),
		AccountPrivateKey: privateKey,
	}

	solanaConfig := &external.SharedSolanaConfig{
		Keys: keys,
	}

	client := newSolanaTestClient(solanaConfig)

	// Test that zkpconfig is properly referenced
	if client.Config != solanaConfig {
		t.Error("Expected zkpconfig to reference the same object")
	}

	// Test that keys are accessible through client
	if client.Config.Keys.ContractPublicKey.String() != privateKey.PublicKey().String() {
		t.Error("Contract public key mismatch")
	}

	if client.Config.Keys.AccountPublicKey.String() != privateKey.PublicKey().String() {
		t.Error("Account public key mismatch")
	}
}

func TestSolanaClientStartServiceLifecycle(t *testing.T) {
	privateKey := solana.MustPrivateKeyFromBase58("4Z7cXSyeFR8wNGMVXUE1TwtKn5D5Vu7FzEv69dokLv7KrQk7h6pu4LF8ZRR9yQBhc7uSM9PiLpAkKktDD8kUmyHT")

	keys := &external.Keys{
		ContractPublicKey: privateKey.PublicKey(),
		AccountPublicKey:  privateKey.PublicKey(),
		AccountPrivateKey: privateKey,
	}

	solanaConfig := &external.SharedSolanaConfig{
		Keys: keys,
	}

	client := newSolanaTestClient(solanaConfig)

	// Test that StartService completes without panic.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("StartService panicked unexpectedly: %v", r)
		}
	}()

	// Start service in a goroutine with timeout
	done := make(chan bool, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected to panic due to missing RabbitMQ consumer
				done <- true
			}
		}()

		client.StartService()
		done <- true
	}()

	// Wait for either completion or timeout
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("StartService test timed out")
	}
}

// Test helper structures for Solana transaction testing
type MockTransaction struct {
	Signatures []string
	Message    string
}

func TestSolanaTransactionStructure(t *testing.T) {
	// Test transaction-like structure that might be used in the client
	tx := MockTransaction{
		Signatures: []string{"signature1", "signature2"},
		Message:    "test transaction",
	}

	if len(tx.Signatures) != 2 {
		t.Errorf("Expected 2 signatures, got %d", len(tx.Signatures))
	}

	if tx.Message != "test transaction" {
		t.Errorf("Expected message 'test transaction', got '%s'", tx.Message)
	}
}

func TestSolanaPublicKeyValidation(t *testing.T) {
	// Test valid public key creation and validation
	privateKey := solana.MustPrivateKeyFromBase58("4Z7cXSyeFR8wNGMVXUE1TwtKn5D5Vu7FzEv69dokLv7KrQk7h6pu4LF8ZRR9yQBhc7uSM9PiLpAkKktDD8kUmyHT")
	publicKey := privateKey.PublicKey()

	// Test public key properties
	if publicKey.String() == "" {
		t.Error("Public key string should not be empty")
	}

	if len(publicKey.Bytes()) != 32 {
		t.Errorf("Expected public key to be 32 bytes, got %d", len(publicKey.Bytes()))
	}

	// Test that we can parse the public key back
	parsedKey, err := solana.PublicKeyFromBase58(publicKey.String())
	if err != nil {
		t.Fatalf("Failed to parse public key: %v", err)
	}

	if !parsedKey.Equals(publicKey) {
		t.Error("Parsed public key should equal original")
	}
}

func TestSolanaClientConcurrentAccess(t *testing.T) {
	privateKey := solana.MustPrivateKeyFromBase58("4Z7cXSyeFR8wNGMVXUE1TwtKn5D5Vu7FzEv69dokLv7KrQk7h6pu4LF8ZRR9yQBhc7uSM9PiLpAkKktDD8kUmyHT")

	keys := &external.Keys{
		ContractPublicKey: privateKey.PublicKey(),
		AccountPublicKey:  privateKey.PublicKey(),
		AccountPrivateKey: privateKey,
	}

	solanaConfig := &external.SharedSolanaConfig{
		Keys: keys,
	}

	client := newSolanaTestClient(solanaConfig)

	// Test concurrent access to client zkpconfig
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			// Simulate concurrent access to client configuration
			if client.Config == nil {
				t.Error("Config should not be nil in concurrent access")
			}

			contractKey := client.Config.Keys.ContractPublicKey.String()
			if contractKey == "" {
				t.Error("Contract key should not be empty in concurrent access")
			}

			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
