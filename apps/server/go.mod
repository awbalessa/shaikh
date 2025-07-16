module github.com/awbalessa/shaikh/apps/server

go 1.24.3

replace github.com/awbalessa/shaikh/apps/server => .

require (
	github.com/jackc/pgx/v5 v5.7.5
	github.com/joho/godotenv v1.5.1
	github.com/pgvector/pgvector-go v0.3.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/text v0.25.0 // indirect
)
