# StarVault

StarVault is a high-performance, trust-and-consent layer that sits between applications and a user's personal data. Rather than applications holding raw personal data directly, they request scoped, time-boxed access through StarVault. 

StarVault enforces these consent policies, manages encryption, and writes tamper-evident proofs of every data access attempt to a blockchain public ledger.

## 🌟 Core Functionality

- **Consent Management:** Users grant, manage, and revoke granular, time-bound access to specific scopes of their data (e.g., `medical_data`, `profile`). Applications must request access using this active consent.
- **Bring Your Own Key (BYOK):** Users can optionally provide their own encryption keys. StarVault encrypts their data using this key (via AES-256-GCM), meaning neither the applications nor StarVault itself can read the data without the user's explicit key provision.
- **Immutable Audit Trail:** Every single data access attempt (allowed or denied) is asynchronously logged, hashed, built into a Merkle Tree, and anchored to a blockchain via a smart contract. This provides mathematical proof of all access history.
- **High-Performance & Scalable:** Designed for strict reliability and high load. It utilizes Redis for distributed rate-limiting and read-through caching, PgBouncer for database transaction pooling, and NATS JetStream for decoupled, adaptive-backpressure background processing.

## 🏗️ Architecture

The system is split into specialized microservices optimized for their specific tasks:

1. **API Gateway (Node.js/Express):** Handles all external client traffic. Responsible for TLS termination, distributed rate-limiting (via Redis + Lua), JWT verification, and routing requests to the Core service.
2. **Core Service (Go):** The high-performance engine of StarVault. It handles consent validation, BYOK encryption/decryption, database interactions, and publishes audit events.
3. **Background Workers (Go):** Adaptive workers that consume audit events from NATS JetStream, generate Merkle roots, and anchor them to the blockchain smart contract at dynamically scaling intervals.
4. **Data Layer:**
   - **PostgreSQL:** The primary persistent datastore for consents, user profiles, and apps.
   - **PgBouncer:** Connection pooler preventing Postgres connection starvation under high concurrency.
   - **Redis:** Used for both gateway rate-limiting state and read-through caching for active consents.
   - **NATS JetStream:** A persistent message broker for reliable audit log processing.

## 🚀 Getting Started

### Prerequisites
- Docker and Docker Compose
- Node.js v18+ (for running local tests)
- Go 1.20+ (if developing on the Core service)

### 1. Environment Setup
Create a `.env` file in the root directory. You can copy the structure from `.env.example` if it exists. Required core variables include:
```env
POSTGRES_USER=starvault
POSTGRES_PASSWORD=your_secure_password
POSTGRES_DB=starvault_db
POSTGRES_HOST=pgbouncer
POSTGRES_PORT=6432
REDIS_URL=redis://redis:6379
NATS_URL=nats://nats:4222
CORE_URL=http://core:8080
JWT_SECRET=your_jwt_secret
STARVAULT_MASTER_KEY=32_byte_hex_string_for_master_encryption
```

### 2. Run the Stack
StarVault is containerized. Start the entire infrastructure using Docker Compose:
```bash
docker compose up -d --build
```
This will spin up Postgres, PgBouncer, Redis, NATS, the Go Core service, and the Node.js API Gateway.

## 🧪 How to Check & Test the System

### 1. Automated Test Suite (Gateway)
The most comprehensive way to verify the system is working is to run the automated Jest test suite located in the `gateway` service. This covers Auth, BYOK encryption, Consent Flows, and Rate Limiting.

```bash
cd gateway
npm install
npm run test
```
*Note: Make sure the Docker containers are running before executing the tests, as they test the live gateway endpoints.*

### 2. Manual Verification Workflow

You can manually trace the core flow using `curl` or Postman.

**Step A: Register a User & App**
- `POST http://localhost:3000/api/v1/auth/register` (Create a user)
- `POST http://localhost:3000/api/v1/apps/register` (Create an application)

**Step B: Grant Consent**
- `POST http://localhost:3000/api/v1/consents` 
- Provide the `appId`, requested `scopes` (e.g., `["medical_data"]`), and an `expires_at` date.

**Step C: Provide BYOK Key (Optional)**
- `POST http://localhost:3000/api/v1/user/key`
- Submit a 32-byte hex string to encrypt the user's vault.

**Step D: Access Data**
- `POST http://localhost:3000/api/v1/access/data`
- The application requests data. The Gateway routes to Core.
- Core checks Redis cache for consent. If active, it decrypts the data using the provided BYOK key (or master key).
- An access event is fired to NATS.

**Step E: Check Audit Worker Logs**
To verify the blockchain worker is adapting and processing logs, view the core container logs:
```bash
docker compose logs -f core
```
You should see messages like:
`BatchWorker: COMMITTED batch <id> (X events, root: <hash>...)` and `Anchored batch on-chain. TxHash: <hash>`.

## 📁 Repository Structure

```text
.
├── core/                   # Go microservice (Consent, Crypto, Audits)
│   ├── internal/           # Handlers, Repositories, Services, Cache
│   ├── main.go             # Core application entrypoint
│   └── Dockerfile          # Core container definition
├── gateway/                # Node.js API Gateway
│   ├── src/                # Middleware (Rate limits, Auth), Routing
│   ├── tests/              # Jest integration tests
│   └── Dockerfile          # Gateway container definition
├── db/
│   └── migrations/         # PostgreSQL schema definitions
├── docker-compose.yml      # Infrastructure orchestration
└── README.md
```
