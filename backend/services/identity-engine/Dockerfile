# NOTE: build context must be the backend/ root, because this module references
# the generated stubs via `replace ... => ../../gen/go`.
#   docker build -f services/identity-engine/Dockerfile -t identity-engine .
# (docker-compose sets `context: .` + this dockerfile.)
FROM golang:1.23 AS build
WORKDIR /src

# Shared generated stubs module (resolved by the replace directive).
COPY gen/go ./gen/go
# Service module.
COPY services/identity-engine ./services/identity-engine

WORKDIR /src/services/identity-engine
RUN CGO_ENABLED=0 go build -o /out/identity-engine ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/identity-engine /identity-engine
EXPOSE 50053
# The RS256 private key is provided at runtime (mounted volume / secret),
# never baked into the image. JWT_PRIVATE_KEY_PATH points at it.
ENTRYPOINT ["/identity-engine"]
