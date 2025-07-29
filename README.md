## Getting Started

Follow these steps to set up and run the project locally:

### 1. Smart Contract Setup

Navigate to the `smart-contract` directory and run the following command to verify and download any required tools:

```bash
cd smart-contract
./run.sh check
```

This will ensure that all dependencies (e.g., Solana CLI, Rust toolchain, Anchor) are installed with their latest versions.

---

### 2. Install and Configure Solana CLI

Install the Solana CLI and configure it to use a local node by following the official Solana documentation:
📖 [Solana CLI Installation Guide](https://solana.com/docs/intro/installation#solana-cli-basics)

---

### 3. Run Local Validator & Deploy Smart Contract

Start the local Solana test validator in a separate terminal:

```bash
solana-test-validator
```

Then deploy the smart contract for development:

```bash
./run.sh deploy
```

This will compile the program and deploy it to the local validator.

---

### 4. Configure Wallet

Copy your Solana wallet private key (JSON file) into the `blockchain-client` directory.
This key will be used as the **program owner** and **transaction payer** for development purposes.

---

### 5. Run the Application

Start the full stack environment using Docker Compose:

```bash
docker-compose up --build
```

This will build and launch all necessary services defined in the Docker Compose configuration.

---

## Architecture Overview

The system architecture is illustrated below:

<img width="2048" height="1188" alt="image" src="https://github.com/user-attachments/assets/8fdf2486-eec3-4d33-ae53-743c0872832b" />
