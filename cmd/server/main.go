package main

import (
	"bridge/api/v1/pb"
	"bridge/core/config"
	"bridge/core/db"
	"bridge/core/logger"
	"bridge/core/repository"
	"bridge/core/server"
	"bridge/services/auth"
	"bridge/services/user"
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	var (
		appName = config.Get(config.AppName, "bridge")
		l       = logger.NewLogger().With().Str("app_name", appName).Logger()
	)

	dbConn, err := db.NewConnection()
	if err != nil {
		l.Fatal().Err(err).Msg("db connection failed")
	}

	rs := repository.NewStore()
	rs.UserRepo = user.NewRepo(dbConn)

	var (
		ctx        = context.Background()
		grpcGWPort = config.Get(config.GrpcGWPort, "0.0.0.0:8001")
		grpcPort   = config.Get(config.GrpcPort, ":8000")
		jwtKey     = config.Get[string](config.JWTKey, "")
	)

	jwtManager, err := auth.NewPasetoToken(jwtKey)
	if err != nil {
		l.Fatal().Err(err).Msg("jwt manager initialization failed")
	}

	var (
		authSvc = auth.NewServer(jwtManager, l, rs)
		grpcSrv = server.NewGrpcSrv()
	)

	pb.RegisterAuthServiceServer(grpcSrv, authSvc)

	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to start net listener")
	}

	l.Info().Msgf("starting grpc server on %v", grpcPort)
	go func() {
		if err = grpcSrv.Serve(lis); err != nil {
			l.Fatal().Err(err).Msg("failed to start grpc server")
		}
	}()

	grpcDialOpts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.DialContext(ctx, grpcPort, grpcDialOpts...)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to dial grpc server")
	}

	gmux := runtime.NewServeMux()
	if err = pb.RegisterAuthServiceHandler(ctx, gmux, conn); err != nil {
		l.Fatal().Err(err).Msg("failed to register auth svc gateway")
	}

	gwServer := &http.Server{
		Addr:    grpcGWPort,
		Handler: gmux,
	}

	l.Info().Msgf("starting gRPC-Gateway on %v", gwServer.Addr)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	go func() {
		sig := <-sigChan
		l.Info().Msgf("shutting down servers, received os signal - %v", sig)

		if err = gwServer.Shutdown(ctx); err != nil {
			l.Fatal().Err(err).Msg("failed to stop gRPC-Gateway server")
		}

		grpcSrv.GracefulStop()
	}()

	if err = gwServer.ListenAndServe(); err != nil {
		l.Fatal().Err(err).Msg("failed to start gRPC-Gateway server")
	}
}
