package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository wraps the Queries with transaction support
type Repository struct {
	pool    *pgxpool.Pool
	queries *Queries
}

// NewRepository creates a new repository with the given pool
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool:    pool,
		queries: New(pool),
	}
}

// Queries returns the queries instance for normal (non-transactional) operations
func (r *Repository) Queries() *Queries {
	return r.queries
}

// WithTransaction executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
// Otherwise, the transaction is committed
func (r *Repository) WithTransaction(ctx context.Context, fn func(*Queries) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create queries with transaction
	txQueries := r.queries.WithTx(tx)

	// Execute the provided function
	if err := fn(txQueries); err != nil {
		return err
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
