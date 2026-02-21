# ADR-002: Double-Entry Ledger Model

## Status

Accepted.

## Context

The system needs to track wallet balances for users. There are two common approaches:

1. **Balance field**: store `balance` as a column on the wallet. Update it on every transaction.
2. **Ledger-based**: store individual entries. Compute balance as `SUM(amount)`.

## Decision

Use a ledger-based model. Balance is never stored — it is always computed:

```sql
SELECT COALESCE(SUM(amount), 0) FROM ledgers WHERE wallet_id = $1;
```

Every transaction creates **two** ledger entries that sum to zero (double-entry bookkeeping):

| Operation | User entry | System entry | Sum |
|---|---|---|---|
| TopUp 5000 | +5000 | -5000 | 0 |
| Spend 2000 | -2000 | +2000 | 0 |

A system wallet acts as the counterparty for all operations.

## Rationale

**Auditability**: the ledger is a complete, append-only log of all money movement. You can reconstruct the balance at any point in time by summing entries up to that timestamp. With a balance field, you only know the current state.

**Correctness under concurrency**: there is no read-then-write race on a balance field. The balance is a derived value from immutable append-only entries. Combined with `FOR UPDATE` locking, this eliminates lost-update problems.

**Self-verifying**: if the system is correct, `SUM(amount)` across ALL ledgers for a given transaction equals zero. This invariant can be verified at any time via a single query:

```sql
SELECT transaction_id, SUM(amount) AS net
FROM ledgers
GROUP BY transaction_id
HAVING SUM(amount) != 0;
```

If this returns rows, the system has a bug.

**No balance drift**: a stored balance field can drift from reality if any code path mutates it incorrectly. A computed balance cannot drift — it is always the truth.

## Consequences

- **Performance**: balance requires a `SUM()` aggregation on every read. For wallets with many entries, this could become slow. Mitigation: `idx_ledger_wallet` index exists. If this becomes a problem, a materialized view or periodic balance snapshot could be added without changing the core model.

- **Storage**: the ledgers table grows proportionally to transaction volume. Each transaction creates 2 rows. There is no archival or pruning strategy.

- **Bonus transactions**: the signup bonus (`BONUS` type, +1000 UC) creates only one ledger entry (credit to user), not two. This breaks the double-entry invariant for that specific transaction type. The system wallet is not debited for bonuses. Whether this is intentional or an oversight is unclear.

## Related

- [ADR-003: Concurrency Strategy](003-concurrency-strategy.md) — how concurrent access to the ledger is serialized.
