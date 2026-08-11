# StarVault Production Deployment Guide

## System Architecture

StarVault is deployed using Docker Compose with three primary components:
1. **Gateway** (Node.js): Exposes the public API, handles rate limiting, and forwards valid requests to Core.
2. **Core** (Go): Contains the business logic, performs encryption/decryption, and interfaces with the database.
3. **Database** (PostgreSQL): Stores user accounts, encrypted vault data, and audit logs.

### Network Isolation

For security, only the **Gateway** service is mapped to a host port (port 3000). The **Core** (port 8080) and **Database** (port 5432) are entirely isolated on an internal Docker network (`starvault_network`). This prevents direct external access to the database or internal core APIs.

## Environment Configuration

Before deploying, copy the `.env.example` file to `.env` and fill in the required production values:

```bash
cp .env.example .env
```

**CRITICAL**: `STARVAULT_MASTER_KEY` must be a cryptographically secure 32-byte hex-encoded string. It must never be checked into version control. 
In a fully hardened production environment, consider sourcing this key dynamically from an external KMS (Key Management Service) or HSM.

## Database Management

### Migrations

StarVault uses the native `docker-entrypoint-initdb.d` directory to initialize the database schema on first run.

- SQL scripts located in `db/migrations/` are automatically executed in alphabetical order when a fresh database container is created.
- For ongoing migrations after initial launch, you should run SQL scripts manually or introduce a migration tool such as `golang-migrate`.

### Backup and Restore

Database backups are performed using the standard `pg_dump` tool against the running PostgreSQL container.

#### Creating a Backup

```bash
docker exec -t <db_container_name> pg_dump -U <postgres_user> -d <postgres_db> -c > backup_$(date +%Y%m%d).sql
```
*Note: This creates a plaintext SQL dump. Ensure backups are stored securely (e.g. encrypted at rest).*

#### Restoring from Backup

To restore data into a clean or existing database container:

```bash
cat backup_filename.sql | docker exec -i <db_container_name> psql -U <postgres_user> -d <postgres_db>
```

## Running the Application

1. Build the production multi-stage images:
   ```bash
   docker-compose build --no-cache
   ```
2. Start the services in detached mode:
   ```bash
   docker-compose up -d
   ```
3. Verify logs to ensure services started correctly:
   ```bash
   docker-compose logs -f
   ```
