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

### 🚀 Start the system

Start the setup script. It will prepare the environment, keep `solana-test-validator` running, and get `docker-compose` up. To stop, use CTRL+C.

```sh
chmod +x setup.sh
```
```sh
./setup.sh
```

More details in [DETAILED START](DETAILED_START.md).

## 🧪 Testing

The project includes comprehensive test suites:

```sh
# Run all unit tests
dev_tools/scripts/test_runner.sh test
```
```sh
# Run benchmarks
dev_tools/scripts/test_runner.sh bench
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
