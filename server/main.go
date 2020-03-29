package main

import (
	"log"
	"net"
	"time"

	"server/config"
	"server/database"
	"server/database/postgres"
	"server/logger"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.uber.org/zap"
)

func main() {
	var (
		err   error
		db    database.Database
		group = errgroup.Group{}
		logg  *zap.Logger
	)

	if err = config.LoadConfig(); err != nil {
		log.Fatal(err)
	}

	if logg, err = logger.SetupLogger(); err != nil {
		log.Fatal(err)
	}
	defer func() { _ = logg.Sync() }()

	zap.ReplaceGlobals(logg)

	if db, err = postgres.New(config.GetConfig().RWDatabase); err != nil {
		log.Fatal(err)
	}

	if err = database.RegisterDatabase(db, false); err != nil {
		log.Fatal(err)
	}
	defer func() { _ = database.Close() }()

	group.Go(func() error { return initializeHTTP(logg) })
	group.Go(func() error { return initializeGRPC(logg) })

	if err = group.Wait(); err != nil {
		log.Fatal(err)
	}
}

func initializeHTTP(logg *zap.Logger) error {
	var (
		r   *gin.Engine
		cfg config.Config = config.GetConfig()
	)

	r = gin.Default()
	r.Use(ginzap.Ginzap(logg, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logg, true))

	v1 := r.Group("v1")
	v1.GET("")

	if cfg.Production {
		gin.SetMode(gin.ReleaseMode)
	}

	return r.Run(cfg.HTTPAddress)
}

func initializeGRPC(logg *zap.Logger) error {
	var (
		cfg config.Config = config.GetConfig()
	)

	listener, err := net.Listen("tcp", cfg.GRPCAddress)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_zap.UnaryServerInterceptor(logg),
				grpc_recovery.UnaryServerInterceptor(
					grpc_recovery.WithRecoveryHandler(logger.PanicRecovery),
				),
			),
		),
	)

	if !cfg.Production {
		log.Println("RUNNING IN DEVELOPMENT REFLECTION ENABLED")
		reflection.Register(grpcServer)
	}

	log.Println("Listening on", cfg.GRPCAddress)
	return grpcServer.Serve(listener)
}
