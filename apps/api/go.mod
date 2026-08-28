module dboke-api

go 1.25.6

require (
	github.com/go-chi/chi/v5 v5.3.2
	github.com/go-chi/cors v1.2.2
	github.com/jackc/pgx/v5 v5.10.0
	github.com/joho/godotenv v1.5.1
	dboke-plugins v0.0.0
)

replace dboke-plugins => ../../plugins

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/go-sql-driver/mysql v1.10.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/hashicorp/go-hclog v1.6.3 // indirect
	github.com/hashicorp/go-plugin v1.8.0 // indirect
	github.com/hashicorp/yamux v0.1.2 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/oklog/run v1.1.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.mongodb.org/mongo-driver v1.17.9 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/net v0.38.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231106174013-bbf56f31fb17 // indirect
	google.golang.org/grpc v1.61.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)
