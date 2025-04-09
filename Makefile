include .env
MIGRATIONS_PATH=./cmd/migrate/migrations

.PHONY: migrate-create
migration:
	@migration create -seq -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-up
migration:
	@migration -path=$(MIGRATIONS_PATH) -database$(DB_MIGRATOR_ADDR) up
	
.PHONY: migrate-down
migration:
	@migration -path=$(MIGRATIONS_PATH) -database$(DB_MIGRATOR_ADDR) down $(filter-out $@,$(MAKECMDGOALS))

	