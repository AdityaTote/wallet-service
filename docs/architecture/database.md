# Database Schema

## Tables

### users

| Column | Type | Constraints |
|---|---|---|
| `id` | `UUID` | PK, `DEFAULT gen_random_uuid()` |
| `username` | `TEXT` | `NOT NULL`, `UNIQUE` |
| `password` | `TEXT` | `NOT NULL` (bcrypt hash) |
| `created_at` | `TIMESTAMPTZ` | `DEFAULT now()` |

### assets

Defines currency types. Currently only `UC` (Universal Credits) exists.

| Column | Type | Constraints |
|---|---|---|
| `id` | `UUID` | PK, `DEFAULT gen_random_uuid()` |
| `code` | `TEXT` | `NOT NULL`, `UNIQUE` |
| `name` | `TEXT` | `NOT NULL` |
| `created_at` | `TIMESTAMPTZ` | `DEFAULT now()` |

### wallets

Links an owner (USER or SYSTEM) to an asset. One wallet per owner per asset.

| Column | Type | Constraints |
|---|---|---|
| `id` | `UUID` | PK, `DEFAULT gen_random_uuid()` |
| `owner_type` | `wallet_owner_type` | `NOT NULL` |
| `owner_id` | `UUID` | `NOT NULL`, `UNIQUE` |
| `asset_id` | `UUID` | `NOT NULL`, FK → `assets(id)` |
| `created_at` | `TIMESTAMPTZ` | `DEFAULT now()` |

Unique constraint: `(owner_type, owner_id, asset_id)`.

### transactions

Each record represents one atomic operation (topup, spend, or bonus). The `id` is **client-supplied** — not auto-generated — to support idempotency.

| Column | Type | Constraints |
|---|---|---|
| `id` | `UUID` | PK (client-supplied) |
| `type` | `transaction_type` | `NOT NULL` |
| `created_at` | `TIMESTAMPTZ` | `DEFAULT now()` |

### ledgers

The core bookkeeping table. Each transaction produces two entries (one per wallet) with amounts that sum to zero.

| Column | Type | Constraints |
|---|---|---|
| `id` | `UUID` | PK, `DEFAULT gen_random_uuid()` |
| `amount` | `INTEGER` | `NOT NULL` (positive = credit, negative = debit) |
| `transaction_id` | `UUID` | `NOT NULL`, FK → `transactions(id)` `ON DELETE CASCADE` |
| `wallet_id` | `UUID` | `NOT NULL`, FK → `wallets(id)` |
| `created_at` | `TIMESTAMPTZ` | `DEFAULT now()` |

Unique constraint: `(transaction_id, wallet_id)` — one entry per wallet per transaction.

## Enum Types

```sql
CREATE TYPE transaction_type AS ENUM ('SPEND', 'TOPUP', 'BONUS');
CREATE TYPE wallet_owner_type AS ENUM ('USER', 'SYSTEM');
```

## Indexes

| Index | Table | Columns | Purpose |
|---|---|---|---|
| `idx_wallets_owner` | `wallets` | `(owner_type, owner_id)` | Wallet lookup by owner |
| `idx_wallets_asset` | `wallets` | `(asset_id)` | Wallet lookup by asset |
| `idx_transactions_created_at` | `transactions` | `(created_at DESC)` | Time-ordered transaction listing |
| `idx_ledger_wallet` | `ledgers` | `(wallet_id)` | Balance calculation: `SUM(amount) WHERE wallet_id = ?` |
| `idx_ledger_tnx` | `ledgers` | `(transaction_id)` | Entries by transaction |
| `idx_ledger_wallet_tnx` | `ledgers` | `(wallet_id, transaction_id)` | Composite lookup, uniqueness |
| `idx_ledger_created_at` | `ledgers` | `(created_at DESC)` | Time-ordered history |

## Entity Relationships

```
users ──1:1──▶ wallets ◀──N:1── ledgers ──N:1──▶ transactions
                  │
                  └── FK ──▶ assets
```

- Each user has one wallet (enforced by `owner_id UNIQUE` on wallets).
- One system wallet exists (owner_type = `SYSTEM`).
- Each transaction produces exactly two ledger entries (user wallet + system wallet).
- Balance is computed, never stored: `SELECT COALESCE(SUM(amount), 0) FROM ledgers WHERE wallet_id = $1`.

## Double-Entry Example

TopUp of 5000 for user alice:

| ledger entry | wallet | amount | net |
|---|---|---|---|
| 1 | alice | +5000 | |
| 2 | system | -5000 | |
| | | **total** | **0** |

Spend of 2000 by user alice:

| ledger entry | wallet | amount | net |
|---|---|---|---|
| 1 | alice | -2000 | |
| 2 | system | +2000 | |
| | | **total** | **0** |

See [ADR-002](../decisions/002-double-entry-ledger.md) for the rationale behind this design.

## Migrations

Managed by `golang-migrate`. Files in `migrations/`:

| File | Direction |
|---|---|
| `20260220102704_wallet_app_schema.up.sql` | Creates all enums, tables, indexes |
| `20260220102704_wallet_app_schema.down.sql` | Empty (not implemented) |

The down migration being empty means there is no automated rollback. To undo the schema, you would need to drop the tables manually.
