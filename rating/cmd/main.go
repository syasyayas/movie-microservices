package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
	"moviedata.com/gen"
	"moviedata.com/pkg/discovery"
	"moviedata.com/pkg/discovery/consul"
	"moviedata.com/rating/internal/controller/rating"
	grpchandler "moviedata.com/rating/internal/handler/grpc"
	"moviedata.com/rating/internal/repository/mysql"
)

const serviceName = "rating"

func main() {

	logger, _ := zap.NewProduction()
	logger = logger.With(zap.String("service", serviceName))
	defer logger.Sync()

	f, err := os.Open("base.yaml")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var cfg serviceConfig

	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		logger.Fatal("Failed to open config", zap.Error(err))
	}

	logger.Info("Starting the rating service", zap.Int("port", cfg.APIConfig.Port))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	registry, err := consul.NewRegistry("consul:8500")
	if err != nil {
		logger.Fatal("Failed to connect to consul registry", zap.Error(err))
	}

	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("rating:%d", cfg.APIConfig.Port)); err != nil {
		logger.Fatal("Failed to register rating service in consul", zap.Error(err))
	}

	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				logger.Error("Failed to report healthy state", zap.Error(err))
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)

	repo, err := mysql.New(logger)
	if err != nil {
		log.Fatal("Failed to establish mySQL connection", zap.Error(err))
	}
	ctrl := rating.New(repo, nil, logger)
	h := grpchandler.New(ctrl, logger)
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.APIConfig.Port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	reflection.Register(srv)
	gen.RegisterRatingServiceServer(srv, h)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s := <-sigChan
		cancel()
		logger.Info("Received stop signal, gracefully stopping", zap.String("signal", s.String()))
		srv.GracefulStop()
		logger.Info("Gracefully stopped gRPC server")
	}()

	if err := srv.Serve(lis); err != nil {
		logger.Fatal("Failed to start grpc server", zap.Error(err))
	}
	wg.Wait()
}
