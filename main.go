package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	ecpb "google.golang.org/grpc/examples/features/proto/echo"
	hwpb "google.golang.org/grpc/examples/helloworld/helloworld"
)

var port = flag.Int("port", 443, "the port to serve on")

// hwServer is used to implement helloworld.GreeterServer.
type hwServer struct {
	hwpb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *hwServer) SayHello(ctx context.Context, in *hwpb.HelloRequest) (*hwpb.HelloReply, error) {
	log.Printf("Received SayHello request: %v", in)
	return &hwpb.HelloReply{Message: "Hello " + in.Name}, nil
}

type ecServer struct {
	ecpb.UnimplementedEchoServer
}

func (s *ecServer) UnaryEcho(ctx context.Context, req *ecpb.EchoRequest) (*ecpb.EchoResponse, error) {
	log.Printf("Received UnaryEcho request: %v", req)
	return &ecpb.EchoResponse{Message: req.Message}, nil
}

// unaryInterceptor is a simple unary interceptor that logs each request.
func unaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	log.Printf("Received request: %v at method: %s", req, info.FullMethod)
	// Calls the handler
	return handler(ctx, req)
}

// healthCheckHandler handles HTTP requests for health checks.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func main() {
	flag.Parse()

	// Start the HTTP server for liveness and readiness probes
	http.HandleFunc("/livenez", healthCheckHandler)
	http.HandleFunc("/readinez", healthCheckHandler)
	go func() {
		log.Println("Starting HTTP server for health checks on port 8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("failed to start HTTP server: %v", err)
		}
	}()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Printf("server listening at %v\n", lis.Addr())

	s := grpc.NewServer(
		grpc.UnaryInterceptor(unaryInterceptor), // Register the interceptor
	)

	// Register Greeter on the server.
	hwpb.RegisterGreeterServer(s, &hwServer{})

	// Register Echo on the same server.
	ecpb.RegisterEchoServer(s, &ecServer{})

	// Register the health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(s, healthServer)

	// Set the health status to SERVING
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection service on gRPC server.
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
