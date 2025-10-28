# Digital Identity System

## 🎯 Project Overview

The **Digital Identity System** is a cutting-edge blockchain-based platform that leverages **Zero-Knowledge Proofs (ZKP)** to enable privacy-preserving identity verification. Built on the **Solana blockchain**, this system allows users to prove specific attributes about their identity (such as being over 18 years old) without revealing sensitive personal information like their exact birth date.

### 🌟 Key Features

- **Privacy-Preserving Verification**: Prove identity attributes without exposing personal data
- **Zero-Knowledge Proofs**: Uses advanced cryptographic techniques (Groth16 protocol) for secure verification
- **Blockchain Integration**: Stores proof references on Solana for tamper-proof verification
- **Microservices Architecture**: Scalable, distributed system with multiple specialized services
- **Real-time Processing**: Asynchronous message queuing for efficient proof generation and verification

### 🎯 Use Cases

- **Age Verification**: Prove you're over 18 without revealing your exact age
- **Identity Authentication**: Verify identity attributes for KYC compliance
- **Privacy-Compliant Systems**: Build applications requiring identity verification while maintaining user privacy
- **Decentralized Identity**: Create self-sovereign identity solutions

### 🔐 Security & Privacy

- **Zero-Knowledge Architecture**: No sensitive data is exposed during verification
- **Blockchain Immutability**: Proof references are stored on Solana for tamper-proof verification
- **Cryptographic Security**: Uses industry-standard Groth16 zero-knowledge proof system
- **Decentralized Storage**: No central authority controls or stores personal data`

## ✈️ Getting Started

Follow these steps to set up and run the project locally:

### 📋 Prerequisites

Before you begin, ensure you have the following installed on your system:

- **Docker** and **Docker Compose** (for containerized deployment)
- **Rust** (with Cargo)
- **Solana CLI** tools (the latest version)
- **Go** 

```sh
docker --version
docker-compose --version
rustc --version
cargo --version
solana --version
go version
```

This will:
- Compile the Rust smart contract to BPF bytecode
- Deploy it to your local Solana validator
- Display the program ID for use by the blockchain client
- Copy program keypair into the blockchain-client directory

---

### 5. 🔑 Configure Wallet and Keys

The blockchain client requires Solana keypairs for operation:

```bash
cd system/blockchain-client

# 1. Your Solana account keypair (for transaction signing and payment)
cp ~/.config/solana/id.json ./id.json

# Fund your account with test SOL (for transaction fees)
solana airdrop 2 $(solana-keygen pubkey ./id.json)
```

**Important**: These keypairs are used for:
- **identity_app-keypair.json**: Smart contract program ID and ownership
- **id.json**: Transaction signing and fee payment

---

### 6. 🐳 Run the Complete System

Start all services using Docker Compose:

```bash
# From the project root directory
docker-compose up --build
```

This will launch:
- **API Service** (port 8080): Identity management REST API
- **Blockchain Client** (port 8001): ZKP generation and Solana integration
- **RabbitMQ** (port 5672): Message queue for inter-service communication
- **Reverse Proxy** (port 9000): Nginx proxy for load balancing

---

### 7. 🧪 Verify Installation

Test that everything is working correctly:

```bash
# Check API health
curl http://localhost:9000/api/health

# Check services are running
docker-compose ps

# Test ZKP generation (if you have test tools)
cd dev_tools/clis/api-test
go run main.go
```
---

## 🧪 Testing

The project includes comprehensive test suites:

```bash
cd dev_tools/scripts/

# Run all unit tests
./test_runner.sh test

# Run benchmarks
./test_runner.sh bench
```
---

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.
e submitting PRs

---

## 🆘 Support

For questions, issues, or contributions:

- **Issues**: [GitHub Issues](https://github.com/bsc-digital-identity/digital-identity-system/issues)
- **Discussions**: [GitHub Discussions](https://github.com/bsc-digital-identity/digital-identity-system/discussions)
- **Documentation**: Check the `/docs` directory for additional documentation
