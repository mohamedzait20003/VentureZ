FROM golang:1.23 AS build
WORKDIR /src

# Shared generated stubs module (resolved by the replace directive).
COPY gen/go ./gen/go

# Service module.
COPY services/api-gateway ./services/api-gateway

WORKDIR /src/services/api-gateway
RUN CGO_ENABLED=0 go build -o /out/api-gateway ./cmd/gateway

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/api-gateway /api-gateway
EXPOSE 8080

# identity-engine's RS256 PUBLIC key is provided at runtime (mounted volume),
# never baked in. JWT_PUBLIC_KEY_PATH points at it.
ENTRYPOINT ["/api-gateway"]
