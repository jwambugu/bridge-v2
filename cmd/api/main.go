package main

import (
	"bridge/api/v1/pb"
	"bridge/internal/config"
	"bridge/internal/config/vault"
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
	"syscall"
)

func main() {
	var (
		appName = config.Get(config.AppName, "bridge")
		ctx     = context.Background()

		l          = logger.NewLogger().With().Str("app_name", appName).Interface("env", config.GetEnvironment()).Logger()
		svcLogger  = l.With().Str("category", "svc").Logger()
		repoLogger = l.With().Str("category", "repo").Logger()

		vaultAddr  = os.Getenv("VAULT_ADDR")
		vaultPath  = os.Getenv("VAULT_PATH")
		vaultToken = os.Getenv("VAULT_TOKEN")
	)

	vaultClient, err := vault.NewProvider(vaultAddr, vaultPath, vaultToken)
	if err != nil {
		l.Fatal().Err(err).Msg("error connecting to vault")
	}

	configProvider := config.NewConfig(vaultClient)

	// getConfig fetches the provided key, if an error occurs, it fails and exits
	getConfig := func(key string) string {
		val, err := configProvider.Get(ctx, key)
		if err != nil {
			l.Fatal().Err(err).Msgf("error fetching key: %s", key)
		}

		return val
	}

	dbConn, err := db.NewConnection(getConfig("DB_DSN"))
	if err != nil {
		l.Fatal().Err(err).Msg("db connection failed")
	}

	rs := repository.NewStore()
	rs.UserRepo = repository.NewUserRepo(dbConn, repoLogger)

	var (
		grpcGWPort = config.Get(config.GrpcGWPort, "0.0.0.0:8001")
		grpcPort   = config.Get(config.GrpcPort, ":8000")
		jwtKey     = getConfig("JWT_SYMMETRIC_KEY")
	)

	jwtManager, err := auth.NewPasetoToken(jwtKey)
	if err != nil {
		l.Fatal().Err(err).Msg("jwt manager initialization failed")
	}

	var (
		unarySrvInterceptors = interceptors.NewUnaryServerInterceptors()
		authSvc              = auth.NewService(jwtManager, svcLogger, rs)
		authProcessor        = auth.NewAuthProcessor(jwtManager, svcLogger, rs)
		grpcSrv              = server.NewGrpcSrv(authProcessor, unarySrvInterceptors)
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
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		l.Info().Msgf("shutting down server, received os signal - %v", sig)

		if err = gwServer.Shutdown(ctx); err != nil {
			l.Fatal().Err(err).Msg("failed to stop gRPC-Gateway server")
		}

		grpcSrv.GracefulStop()
	}()

	if err = gwServer.ListenAndServe(); err != nil {
		l.Fatal().Err(err).Msg("failed to start gRPC-Gateway server")
	}
}
