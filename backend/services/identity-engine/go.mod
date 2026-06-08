module github.com/venturez/backend/services/identity-engine

go 1.23

require (
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/jackc/pgx/v5 v5.6.0
	github.com/venturez/backend/gen/go v0.0.0
	golang.org/x/crypto v0.27.0
	google.golang.org/grpc v1.66.0
)

require (
	github.com/jackc/pgpassfile v1.0.0
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a
	github.com/jackc/puddle/v2 v2.2.1
	golang.org/x/net v0.26.0
	golang.org/x/sync v0.8.0
	golang.org/x/sys v0.25.0
	golang.org/x/text v0.18.0
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240604185151-ef581f913117
	google.golang.org/protobuf v1.34.2
)

replace github.com/venturez/backend/gen/go => ../../gen/go
