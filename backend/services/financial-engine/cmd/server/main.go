package main

import (
	"log"
	"net"
)

// Entrypoint for the financial-engine gRPC server.
// TODO: register the generated FinancialEngineServer (gen/go/financial/v1)
// and wire it to the calc package, then serve on :50051.
func main() {
	const addr = ":50051"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen %s: %v", addr, err)
	}
	log.Printf("financial-engine listening on %s", addr)
	_ = lis
	// grpcServer := grpc.NewServer()
	// financialv1.RegisterFinancialEngineServer(grpcServer, server.New(calc.New()))
	// log.Fatal(grpcServer.Serve(lis))
}
