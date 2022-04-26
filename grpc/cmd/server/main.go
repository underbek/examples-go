package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	pb "github.com/AndreyAndreevich/examples-go/grpc/proto"
	"github.com/AndreyAndreevich/examples-go/grpc/server"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/reflection"
)

func main() {
	srv := server.New()

	// run grpc server
	grpcServer := grpc.NewServer(
		// middlewares
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				// auth
				grpc_auth.UnaryServerInterceptor(func(ctx context.Context) (context.Context, error) {
					fmt.Println(metautils.ExtractIncoming(ctx).Get("token"))
					fmt.Println(metautils.ExtractIncoming(ctx).Get("authorization"))
					return ctx, nil
				}),
			)),
	)

	reflection.Register(grpcServer)

	pb.RegisterUserServiceServer(grpcServer, srv)

	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		grpclog.Fatalf("failed to listen: %v", err)
	}

	go func() {
		log.Fatal(grpcServer.Serve(listener))
	}()

	// run http server
	mux := runtime.NewServeMux()

	pb.RegisterUserServiceHandlerServer(context.Background(), mux, srv)
	go func() {
		log.Fatal(http.ListenAndServe(":8080", mux))
	}()

	// run http gateway
	conn, err := grpc.Dial("localhost:8000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	mux = runtime.NewServeMux(
		// convert header to meta
		runtime.WithIncomingHeaderMatcher(
			func(key string) (string, bool) {
				switch key {
				case "Token":
					return "token", true
				default:
					// parse default headers (example: authorization)
					return runtime.DefaultHeaderMatcher(key)
				}
			},
		),
	)

	client := pb.NewUserServiceClient(conn)
	pb.RegisterUserServiceHandlerClient(context.Background(), mux, client)
	log.Fatal(http.ListenAndServe(":8081", mux))
}
