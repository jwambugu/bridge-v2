package main

import (
	"bridge/api/v1/pb"
	"bridge/internal/config"
	"bridge/internal/db"
	"bridge/internal/interceptors"
	"bridge/internal/logger"
	"bridge/internal/repository"
	"bridge/internal/server"
	"bridge/services/auth"
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	appLogger := logger.NewLogger()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := config.NewDefaultConfig(ctx)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("get default config")
	}

	appLogger = appLogger.With().
		Str("app_name", config.EnvKey.Name).
		Str("env", config.EnvKey.Env).
		Logger()

	var (
		svcLogger  = appLogger.With().Str("category", "svc").Logger()
		repoLogger = appLogger.With().Str("category", "repo").Logger()
	)

	dbConn, err := db.NewConnection(config.EnvKey.DbDsn)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("db connection failed")
	}

	rs := repository.NewStore()
	rs.UserRepo = repository.NewUserRepo(dbConn, repoLogger)

	var (
		grpcGWPort = config.EnvKey.GrpcGatewayPort
		grpcPort   = strconv.Itoa(int(config.EnvKey.Port))
		jwtKey     = config.EnvKey.JwtKey
	)

	jwtManager, err := auth.NewPasetoToken(jwtKey)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("jwt manager initialization failed")
	}

	var (
		unarySrvInterceptors = interceptors.NewUnaryServerInterceptors()
		authSvc              = auth.NewService(jwtManager, svcLogger, rs)
		authProcessor        = auth.NewAuthProcessor(jwtManager, svcLogger, rs)
		grpcSrv              = server.NewGrpcSrv(authProcessor, unarySrvInterceptors)
	)

	pb.RegisterAuthServiceServer(grpcSrv, authSvc)

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("failed to start net listener")
	}

	appLogger.Info().Msgf("starting grpc server on %v", grpcPort)
	go func() {
		if err = grpcSrv.Serve(lis); err != nil {
			appLogger.Fatal().Err(err).Msg("failed to start grpc server")
		}
	}()

	grpcDialOpts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.DialContext(ctx, grpcPort, grpcDialOpts...)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("failed to dial grpc server")
	}

	gmux := runtime.NewServeMux()
	if err = pb.RegisterAuthServiceHandler(ctx, gmux, conn); err != nil {
		appLogger.Fatal().Err(err).Msg("failed to register auth svc gateway")
	}

	gwServer := &http.Server{
		Addr:    grpcGWPort,
		Handler: gmux,
	}

	appLogger.Info().Msgf("starting gRPC-Gateway on %v", gwServer.Addr)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		appLogger.Info().Msgf("shutting down server, received os signal - %v", sig)

		if err = gwServer.Shutdown(ctx); err != nil {
			appLogger.Fatal().Err(err).Msg("failed to stop gRPC-Gateway server")
		}

		grpcSrv.GracefulStop()
	}()

	if err = gwServer.ListenAndServe(); err != nil {
		appLogger.Fatal().Err(err).Msg("failed to start gRPC-Gateway server")
	}
}
