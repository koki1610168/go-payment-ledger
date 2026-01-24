# Go Payment Ledger

A robust payment ledger service written in Go that handles balance management, payments, and transfers with transactional integrity and idempotency guarantees.

## Features

- **Balance Management** — Query client balances with currency support
- **Payments** — Create payments (credits/debits) with atomic balance updates
- **Transfers** — Move funds between clients atomically
- **Idempotency** — Prevent duplicate charges with idempotency keys
- **Transaction Safety** — All operations use database transactions with proper rollback handling
- **Rate Limiting** — Per-IP rate limiting with configurable limits
- **Connection Pooling** — Efficient PostgreSQL connection management with `pgxpool`
- **Ledger History** — Full audit trail of all transactions

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      HTTP Server (:8080)                    │
├─────────────────────────────────────────────────────────────┤
│                    Rate Limiter Middleware                  │
├─────────────────────────────────────────────────────────────┤
│                         Handler                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────────────┐ │
│  │  /payments  │  │  /transfer  │  │ /clients/{id}/...   │ │
│  └─────────────┘  └─────────────┘  └─────────────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                     Store (ClientStore)                     │
├─────────────────────────────────────────────────────────────┤
│                   PostgreSQL (pgxpool)                      │
└─────────────────────────────────────────────────────────────┘
```

## Prerequisites

- Go 1.25+
- PostgreSQL 14+

## Installation

```bash
# Clone the repository
git clone https://github.com/koki1610168/go-payment-ledger.git
cd go-payment-ledger

# Download dependencies
go mod download

# Build the server
go build -o ledger-server ./cmd/server
```

## Configuration

Set the following environment variable before running:

| Variable | Description | Required |
|----------|-------------|----------|
| `DATABASE_URL` | PostgreSQL connection string | Yes |

Example:
```bash
export DATABASE_URL="postgres://user:password@localhost:5432/payment_ledger?sslmode=disable"
```

## Database Schema

Run the following SQL to set up the database:

```sql
-- Clients table: stores account balances
CREATE TABLE clients (
    client_id  TEXT PRIMARY KEY,
    balance    BIGINT NOT NULL DEFAULT 0,
    currency   TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Ledger entries: immutable transaction log
CREATE TABLE ledger_entries (
    entry_id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id       TEXT NOT NULL REFERENCES clients(client_id),
    amount          BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    idempotency_key TEXT
);

-- Index for idempotency lookups
CREATE INDEX idx_ledger_idempotency ON ledger_entries(idempotency_key) WHERE idempotency_key IS NOT NULL;

-- Index for client ledger queries
CREATE INDEX idx_ledger_client ON ledger_entries(client_id);
```

## Running the Server

```bash
# Set the database URL
export DATABASE_URL="postgres://user:password@localhost:5432/payment_ledger?sslmode=disable"

# Run the server
./ledger-server

# Or run directly with go
go run ./cmd/server
```

The server starts on port **8080**.

## API Reference

### Get Balance

Retrieve the current balance for a client.

```http
GET /clients/{clientId}/balance
```

**Response:**
```json
{
  "ClientID": "client_001",
  "Balance": 10000,
  "Currency": "JPY"
}
```

---

### Get Ledger

Retrieve the full transaction history for a client.

```http
GET /clients/{clientId}/ledger
```

**Response:**
```json
{
  "client_id": "client_001",
  "ledger_entires": [
    {
      "EntryId": "550e8400-e29b-41d4-a716-446655440000",
      "ClientId": "client_001",
      "Amount": 1400,
      "CreatedAt": "2026-01-24T10:30:00Z",
      "IdempotencyKey": "pay-001"
    }
  ]
}
```

---

### Create Payment

Create a payment (credit or debit) for a client.

```http
POST /payments
Content-Type: application/json
```

**Request Body:**
```json
{
  "clientID": "client_001",
  "amount": 1400,
  "currency": "JPY",
  "idempotencyKey": "payment-12345"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `clientID` | string | Client identifier (required) |
| `amount` | integer | Amount in smallest currency unit. Positive for credit, negative for debit (required) |
| `currency` | string | Currency code (required) |
| `idempotencyKey` | string | Unique key to prevent duplicate processing (required) |

**Response:**
```json
{
  "ClientID": "client_001",
  "Balance": 11400,
  "Currency": "JPY"
}
```

---

### Transfer Funds

Transfer funds between two clients atomically.

```http
POST /transfer
Content-Type: application/json
```

**Request Body:**
```json
{
  "from_client_id": "client_001",
  "to_client_id": "client_002",
  "amount": 300,
  "idempotencyKey": "transfer-001"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `from_client_id` | string | Source client identifier (required) |
| `to_client_id` | string | Destination client identifier (required) |
| `amount` | integer | Transfer amount, must be positive (required) |
| `idempotencyKey` | string | Unique key to prevent duplicate processing (required) |

**Response:**
```json
{
  "from_client_id": "client_001",
  "to_client_id": "client_002",
  "amount": 300,
  "from_new_balance": 9700,
  "to_new_balance": 10300
}
```

## Error Handling

The API returns appropriate HTTP status codes:

| Status Code | Description |
|-------------|-------------|
| `200 OK` | Request successful |
| `400 Bad Request` | Invalid request body or missing required fields |
| `404 Not Found` | Client not found |
| `405 Method Not Allowed` | Invalid HTTP method |
| `429 Too Many Requests` | Rate limit exceeded (includes `Retry-After` header) |

## Rate Limiting

The server implements per-IP rate limiting using a token bucket algorithm:

- **Rate:** 10 requests per second
- **Burst:** 20 requests

When rate limited, responses include a `Retry-After: 1` header.

## Testing

```bash
# Run unit tests (no database required)
go test ./...

# Run integration tests (requires DATABASE_URL)
export DATABASE_URL="postgres://user:password@localhost:5432/payment_ledger_test?sslmode=disable"
go test -v ./internal/server/...
```

## Design Decisions

### Transaction Safety

All balance-modifying operations use PostgreSQL transactions with:
- `FOR UPDATE` row locks to prevent concurrent modifications
- Proper rollback on any failure via `defer tx.Rollback()`
- Atomic commit ensuring ledger entries and balance updates succeed or fail together

### Idempotency

Every payment and transfer requires an idempotency key. If a request with the same key is received:
- The original result is returned
- No duplicate ledger entries are created
- Balance remains unchanged after the first successful request

This prevents double-charging in scenarios like network retries or client failures.

### Amount Representation

Amounts are stored as `BIGINT` (int64) representing the smallest currency unit:
- JPY: 1 = ¥1
- USD: 100 = $1.00

This avoids floating-point precision issues common in financial applications.

### Connection Pooling

Database connections are managed via `pgxpool` with:
- Max 10 concurrent connections
- Min 1 idle connection
- 5-second idle timeout
- 30-second connection lifetime

## Project Structure

```
go_payment_ledger/
├── cmd/
│   └── server/
│       └── main.go          # Application entrypoint
├── internal/
│   └── server/
│       ├── db.go            # Database connection management
│       ├── handler.go       # HTTP handlers and routing
│       ├── middleware.go    # Rate limiting middleware
│       ├── store.go         # Data access layer
│       └── *_test.go        # Test files
├── go.mod
├── go.sum
└── README.md
```

## License

MIT License
