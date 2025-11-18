# DETAILED START

### 1. 🛠️ Smart Contract Setup

Navigate to the smart contract directory and verify your development environment:

```bash
cd dev_tools/scripts/
chmod +x smart_contract.sh
smart_contract.sh check
```

This command will:
- Verify Rust toolchain installation
- Check Solana CLI availability
- Display version information for all tools

If any tools are missing, the script will provide installation instructions.

---

### 2. 🔗 Install and Configure Solana CLI

Install the Solana CLI and configure it for local development.

📖 For detailed installation instructions, visit: [Solana CLI Installation Guide](https://solana.com/docs/intro/installation#solana-cli-basics)

---

### 3. 🏃‍♂️ Start Local Solana Validator

```bash
# Start the validator (keep this running in a separate process)
solana-test-validator --reset
```

The validator should be running on `http://localhost:8899` and `ws://localhost:8900`.

---

### 4. 🚀 Deploy Smart Contract

Build and deploy the smart contract to your local validator:

```bash
cd dev_tools/scripts/
smart_contract.sh deploy
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
cp ~/.zkpconfig/solana/id.json ./id.json

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

---
Solana setup instructions:

Requirements:
cargo 1.75.0
solana solana-cli 1.18.26

Steps:
0. solana-test-validator --reset
1. deploy returns <KEYPAIR PATH>, eg: C:\Users\userabc\.config\solana\id.json
2. solana program show --programs
3. set PROGRAM_ID=YOUR_PROGRAM_ID w docker-compose.yml -> services -> blockchain-client -> environment
4. set     
   volumes:
    - "KEYPAIR PATH:/app/id.json:ro"
4. docker-compose build, docker-compose up
5. uv run python main.py

