# API Reference

Base URL: `http://localhost:8080/api`

All responses use this envelope:

```json
{"success": true|false, "message": "...", "data": ...}
```

`data` is omitted on errors.

## Authentication

Wallet endpoints require:

```
Authorization: Bearer <access_token>
```

Tokens are obtained from `/api/auth/signup` or `/api/auth/signin`. They are HS256 JWTs with 24h TTL, containing `uid` (user UUID) and `wid` (wallet UUID).

The auth middleware extracts the token, verifies the signature, looks up the user and wallet in the database, and injects them into the request context. If any step fails, the response is `401 {"success": false, "message": "unauthorized"}`.

---

## Endpoints

### GET /api/health

No authentication required.

Returns service status and database connectivity check.

```bash
curl http://localhost:8080/api/health
```

```json
{
  "success": true,
  "message": "service is healthy",
  "data": {
    "status": "healthy",
    "timestamp": "2026-02-21T08:00:00Z",
    "checks": {
      "database": {
        "status": "healthy",
        "response_time": "1.2ms"
      }
    }
  }
}
```

---

### POST /api/auth/signup

No authentication required.

Creates a user, a wallet (linked to the `UC` asset), and grants a 1000 UC signup bonus via a `BONUS` transaction.

**Request:**

```json
{"username": "string (required)", "password": "string (required)"}
```

**Response (201):**

```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "id": "uuid",
    "username": "string",
    "wallet_id": "uuid",
    "balance": 1000,
    "access_token": "jwt-string"
  }
}
```

**Errors:**

| Status | Cause |
|---|---|
| 400 | Malformed JSON or unknown fields in body |
| 422 | Missing `username` or `password` |
| 500 | Username already taken, or internal failure |

---

### POST /api/auth/signin

No authentication required.

**Request:**

```json
{"username": "string (required)", "password": "string (required)"}
```

**Response (200):**

```json
{
  "success": true,
  "message": "User logged in successfully",
  "data": {
    "id": "uuid",
    "username": "string",
    "access_token": "jwt-string"
  }
}
```

Note: signin does not return `wallet_id` or `balance`. These are embedded in the JWT claims and available via the balance endpoint.

**Errors:**

| Status | Cause |
|---|---|
| 400 | Malformed JSON |
| 422 | Missing fields |
| 500 | Wrong credentials or internal failure |

---

### POST /api/wallet/topup

**Requires auth.**

Adds funds to the user's wallet. Creates a `TOPUP` transaction with two ledger entries: `+amount` on user wallet, `-amount` on system wallet.

**Request:**

```json
{
  "txn_id": "uuid (required, client-generated)",
  "amount": 5000
}
```

`amount` must be a positive integer.

**Response (201):**

```json
{"success": true, "message": "wallet topped up successfully", "data": 6000}
```

`data` is the new balance (integer).

**Idempotent retry:** If `txn_id` already exists, returns `201` with `"message": "transaction with id already exist"` and the current balance. No duplicate entries are created.

**Errors:**

| Status | Cause |
|---|---|
| 400 | Missing/invalid `txn_id`, missing `amount`, `amount` ≤ 0 |
| 401 | Missing or invalid token |
| 500 | Transaction failure |

---

### POST /api/wallet/spend

**Requires auth.**

Deducts funds from the user's wallet. Creates a `SPEND` transaction with two ledger entries: `-amount` on user wallet, `+amount` on system wallet. Checks balance before proceeding.

**Request:**

```json
{
  "txn_id": "uuid (required, client-generated)",
  "amount": 2000
}
```

**Response (201):**

```json
{"success": true, "message": "wallet spend up successfully", "data": 4000}
```

Same idempotent retry behavior as topup.

**Errors:**

| Status | Cause |
|---|---|
| 400 | Missing/invalid fields, `amount` ≤ 0 |
| 401 | Missing or invalid token |
| 500 | Insufficient balance, or transaction failure |

Note: insufficient balance currently returns 500, not 400. The `ErrInsufficientBalance` error is defined with status 400 in `models/error.go`, but it is wrapped by a generic 500 in `service/wallet.go`. This is a known inconsistency in error handling.

---

### GET /api/wallet/balance

**Requires auth.**

Returns the current balance for the authenticated user's wallet.

```bash
curl http://localhost:8080/api/wallet/balance \
  -H "Authorization: Bearer <token>"
```

**Response (200):**

```json
{"success": true, "message": "wallet balance retrieved successfully", "data": 6000}
```

`data` is the balance (integer), computed as `SUM(amount)` from all ledger entries for this wallet.

---

## Input Validation

- JSON bodies must not contain unknown fields (`DisallowUnknownFields` is enabled).
- `username` and `password` are validated as required strings.
- `txn_id` must be a valid UUID.
- `amount` must be a positive integer (`gt=0`).

Validation errors return structured messages like `"amount is required"` or `"amount must be greater than 0"`.
