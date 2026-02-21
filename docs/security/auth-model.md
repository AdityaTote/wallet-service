# Security & Auth Model

## Authentication Flow

1. User calls `POST /api/auth/signup` or `POST /api/auth/signin` with `{username, password}`.
2. Server validates credentials, returns a JWT access token.
3. Client includes the token in subsequent requests: `Authorization: Bearer <token>`.
4. Auth middleware validates the token on protected routes (`/api/wallet/*`).

## JWT Tokens

| Property | Value |
|---|---|
| Algorithm | HS256 (HMAC-SHA256) |
| TTL | 24 hours |
| Issuer | `wallet-service` |
| Signing key | `JWT_SECRET` environment variable |

### Claims

```json
{
  "uid": "user-uuid-string",
  "wid": "wallet-uuid",
  "exp": 1234567890,
  "iat": 1234567890,
  "iss": "wallet-service"
}
```

`uid` is stored as a string (not UUID) in the token. `wid` is a UUID. The middleware parses `uid` back into a UUID after extracting it.

### Token Validation (Middleware)

The auth middleware (`internal/middleware/auth.go`) performs these steps:

1. Extract `Authorization` header
2. Split on space, expect exactly `Bearer <token>`
3. Parse and verify JWT signature against `JWT_SECRET`
4. Parse `uid` claim into UUID
5. Query database: `GetUserById(uid)` — verify user exists
6. Query database: `GetWalletByOwner(user.ID)` — get wallet
7. Inject `models.User{Id, WalletId}` into request context

Steps 5-6 mean every authenticated request performs 2 database queries. There is no caching.

## Password Storage

- Hashing: `bcrypt.GenerateFromPassword` with `bcrypt.DefaultCost` (cost factor 10)
- Verification: `bcrypt.CompareHashAndPassword`
- Passwords are never logged or returned in responses

## Secrets Management

The `JWT_SECRET` is loaded from:
1. `.env` file (local development)
2. Environment variable (production)

There is no secrets rotation mechanism. Changing `JWT_SECRET` invalidates all existing tokens immediately.

The `.env.example` contains a placeholder: `JWT_SECRET=your_secret_key_here_use_a_long_random_string`.

## Input Validation

- JSON decoder rejects unknown fields (`DisallowUnknownFields()`)
- Struct validation via `go-playground/validator` with `required` and `gt=0` tags
- `txn_id` is validated as a UUID (via Go's `uuid.UUID` type in JSON unmarshalling)

## What Is Not Implemented

- **HTTPS/TLS**: the HTTP server has no TLS configuration. TLS termination would need to happen at a reverse proxy or load balancer.
- **Rate limiting**: no request throttling. A malicious or buggy client can make unlimited requests.
- **CORS**: no CORS headers are configured.
- **Refresh tokens**: only access tokens exist. When they expire after 24h, the user must re-authenticate.
- **Token revocation**: there is no blacklist. A compromised token is valid until expiration.
- **Password requirements**: no minimum length, complexity, or breach-check validation.
- **Brute force protection**: no account lockout or exponential backoff on failed auth attempts.
- **Audit logging**: financial operations are not logged in a structured audit log (only zerolog debug/error entries).

## Context Key

The middleware injects user data using the string key `"user"`:

```go
ctx := context.WithValue(r.Context(), "user", models.User{...})
```

This uses a plain string as a context key, which risks collisions. The Go convention is to use an unexported key type. This is a minor code quality issue, not a security vulnerability.
