.PHONY: sqlc-generate sqlc-clean seed

# Include .env file
include .env
export

# Generate all sqlc code
sqlc-generate:
	@echo "Generating repository code..."
	sqlc generate
	@echo "✓ Done! All code generated in internal/repository/"

# Run database migrations
migrate-up:
	@echo "Running database migrations..."
	migrate -path migrations -database "postgresql://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(DATABASE_HOST):$(DATABASE_PORT)/$(DATABASE_NAME)?sslmode=disable" up
	@echo "✓ Migrations complete!"

# Seed the database with initial data
seed:
	@echo "Seeding database..."
	go run cmd/seed/main.go
	@echo "✓ Database seeded successfully!"

# Run the wallet service
run:
	@echo "Starting wallet service..."
	go run cmd/wallet-service/main.go
