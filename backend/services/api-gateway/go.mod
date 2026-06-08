module github.com/venturez/backend/services/api-gateway

go 1.23

require (
	github.com/go-chi/chi/v5 v5.1.0
	github.com/go-chi/httprate v0.14.1
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/venturez/backend/gen/go v0.0.0
	google.golang.org/grpc v1.66.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240604185151-ef581f913117 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

// Generated stubs live in the monorepo; resolve them locally.
replace github.com/venturez/backend/gen/go => ../../gen/go
