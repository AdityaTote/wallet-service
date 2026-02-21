# ADR-003: Concurrency Strategy

## Status

Accepted.

## Context

Concurrent requests to the same wallet must not produce incorrect results. Specific risks:

- **Lost updates**: two topups running simultaneously both read balance=1000, both add 500, result is 1500 instead of 2000.
- **Double-spending**: two spends check balance=1000, both spend 800, result is -600.
- **Duplicate processing**: network retry sends the same topup twice, credits applied twice.

## Decision

Three mechanisms are used together:

### 1. Row-level locking (`SELECT ... FOR UPDATE`)

Before any mutation, the wallet row is locked within the transaction:

```sql
SELECT id FROM wallets WHERE owner_id = $1 FOR UPDATE;
```

This blocks concurrent transactions targeting the same wallet until the lock-holding transaction completes. Different wallets are never blocked by each other.

### 2. Database transactions

All mutation steps execute within a single PostgreSQL transaction:

```
BEGIN
  SELECT ... FOR UPDATE (lock)
  INSERT transaction
  INSERT ledger entry (user)
  INSERT ledger entry (system)
  SELECT SUM(amount) (read balance)
COMMIT
```

If any step fails, the entire transaction rolls back. The `Repository.WithTransaction` method handles `BEGIN`, deferred `ROLLBACK`, and `COMMIT`.

### 3. Idempotency keys

Every topup/spend request requires a client-supplied `txn_id` (UUID). Before starting a transaction, the service checks if a transaction with that ID exists:

- If it exists: return the current balance without creating new entries.
- If it does not exist: proceed normally.

The `transactions.id` primary key and `ledgers(transaction_id, wallet_id)` unique constraint enforce this at the database level as well.

## Alternatives Considered

| Approach | Why not |
|---|---|
| **Optimistic locking** (version column on wallet) | Requires application-level retry loops. More complex. The ledger model computes balance from entries, so there is no single row to version. |
| **SERIALIZABLE isolation** | PostgreSQL's SERIALIZABLE level detects anomalies and aborts transactions. This would require retry logic for serialization failures. Higher contention under load. |
| **Application-level mutex** (sync.Mutex) | Only works for a single instance. Breaks with horizontal scaling. |
| **Advisory locks** (pg_advisory_lock) | Works, but `FOR UPDATE` is simpler and more idiomatic for row-level access control. |

## Consequences

- **Serialized per-wallet**: concurrent requests to the same wallet are processed sequentially. Throughput for a single highly-active wallet is limited by transaction duration.

- **No deadlocks for single-wallet operations**: each transaction locks exactly one wallet row. Deadlocks would only occur if a single transaction attempted to lock multiple wallets in inconsistent order. The current code locks the user wallet first, then accesses the system wallet — but does NOT lock the system wallet with `FOR UPDATE`. This avoids deadlocks but means the system wallet is not protected from concurrent access. Since the system wallet's balance is not checked (only the user's balance is verified), this is acceptable.

- **pgxpool manages connections**: the connection pool bounds the number of concurrent database connections, providing back-pressure. If all connections are in use, new requests block until a connection is available.

- **Idempotency check happens outside the transaction**: the `txn_id` lookup occurs before `BEGIN`. In a very narrow race window, two identical requests could both pass the check and both attempt to `INSERT INTO transactions`. The second would fail on the primary key constraint and the transaction would roll back. This is safe but the error message would be a generic failure rather than the idempotent response.
