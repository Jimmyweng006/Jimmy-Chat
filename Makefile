migrateup:
	migrate -path db/migration -database "postgresql://postgres:root@localhost:5432/postgres?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://postgres:root@localhost:5432/postgres?sslmode=disable" -verbose down

sqlc:
	sqlc generate