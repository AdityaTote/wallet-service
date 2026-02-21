CREATE TYPE transaction_type AS ENUM ('SPEND', 'TOPUP', 'BONUS');
CREATE TYPE wallet_owner_type AS ENUM ('USER', 'SYSTEM');


CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username TEXT NOT NULL UNIQUE,
  password TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE assets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE wallets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_type wallet_owner_type NOT NULL,
  owner_id UUID NOT NULL UNIQUE,
  asset_id UUID NOT NULL REFERENCES assets(id),
  created_at TIMESTAMPTZ DEFAULT now(),

  UNIQUE (owner_type, owner_id, asset_id)
);

CREATE INDEX idx_wallets_owner ON wallets(owner_type, owner_id);
CREATE INDEX idx_wallets_asset ON wallets(asset_id);

CREATE TABLE transactions (
  id UUID PRIMARY KEY,
  type transaction_type NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_transactions_created_at ON transactions(created_at DESC);

CREATE TABLE ledgers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  amount INTEGER NOT NULL,
  transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
  wallet_id UUID NOT NULL REFERENCES wallets(id),
  created_at TIMESTAMPTZ DEFAULT now(),

  UNIQUE (transaction_id, wallet_id)
);

CREATE INDEX idx_ledger_wallet ON ledgers(wallet_id);
CREATE INDEX idx_ledger_tnx ON ledgers(transaction_id);
CREATE INDEX idx_ledger_wallet_tnx ON ledgers(wallet_id, transaction_id);
CREATE INDEX idx_ledger_created_at ON ledgers(created_at DESC);