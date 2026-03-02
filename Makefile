DB_URL ?=

migrate-up:
	@test -n "$(DB_URL)" || (echo "DB_URL is required"; exit 1)
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	@test -n "$(DB_URL)" || (echo "DB_URL is required"; exit 1)
	migrate -path migrations -database "$(DB_URL)" down
