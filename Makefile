DB_URL=postgres://elibrary:elibrary@localhost:5432/elibrary?sslmode=disable

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down
