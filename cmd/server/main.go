package main

import (
	"bridge/api/v1/pb"
	"bridge/internal/config"
	"bridge/internal/db"
	"bridge/internal/interceptors"
	"bridge/internal/logger"
	"bridge/internal/repository"
	"bridge/internal/servers"
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
	"syscall"
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
		unarySrvInterceptors = interceptors.NewUnaryServerInterceptors()
		authSvc              = auth.NewService(jwtManager, l, rs)
		authProcessor        = auth.NewAuthProcessor(jwtManager, rs)
		grpcSrv              = servers.NewGrpcSrv(authProcessor, unarySrvInterceptors)
	)

	pb.RegisterAuthServiceServer(grpcSrv, authSvc)

	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to start net listener")
	}

	l.Info().Msgf("starting grpc servers on %v", grpcPort)
	go func() {
		if err = grpcSrv.Serve(lis); err != nil {
			l.Fatal().Err(err).Msg("failed to start grpc servers")
		}
	}()

	grpcDialOpts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.DialContext(ctx, grpcPort, grpcDialOpts...)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to dial grpc servers")
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
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		l.Info().Msgf("shutting down servers, received os signal - %v", sig)

		if err = gwServer.Shutdown(ctx); err != nil {
			l.Fatal().Err(err).Msg("failed to stop gRPC-Gateway servers")
		}

		grpcSrv.GracefulStop()
	}()

	if err = gwServer.ListenAndServe(); err != nil {
		l.Fatal().Err(err).Msg("failed to start gRPC-Gateway servers")
	}
}
