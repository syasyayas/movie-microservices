package main

import (
	"context"
	"fmt"
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
	"moviedata.com/metadata/internal/controller/metadata"
	grpchandler "moviedata.com/metadata/internal/handler/grpc"
	"moviedata.com/metadata/internal/repository/memory"
	"moviedata.com/pkg/discovery"
	"moviedata.com/pkg/discovery/consul"
)

const serviceName = "metadata"

func main() {
	logger, _ := zap.NewProduction()
	logger = logger.With(zap.String("service", serviceName))
	f, err := os.Open("base.yaml")
	if err != nil {
		logger.Fatal("Failed to open config file", zap.Error(err))
	}
	defer f.Close()
	var cfg serviceConfig

	registry, err := consul.NewRegistry("consul:8500")
	if err != nil {
		logger.Fatal("Failed to establish consul connection", zap.Error(err))
	}

	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		logger.Fatal("Failed to decode config file", zap.Error(err))
	}
	logger.Info("Starting the movie metadata service on port", zap.Int("port", cfg.APIConfig.Port))

	ctx, cancel := context.WithCancel(context.Background())

	instanceID := discovery.GenerateInstanceID(serviceName)

	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("metadata:%d", cfg.APIConfig.Port)); err != nil {
		logger.Fatal("Faile to register service in consul", zap.Error(err))
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

	repo := memory.New()
	ctrl := metadata.New(repo)
	h := grpchandler.New(ctrl)
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.APIConfig.Port))
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}
	srv := grpc.NewServer()
	reflection.Register(srv)
	gen.RegisterMetadataServiceServer(srv, h)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		s := <-sigChan
		cancel()
		logger.Info("Received signall, attemting graceful shutdown", zap.String("signal", s.String()))
		srv.GracefulStop()
		logger.Info("Gracefully stopped gRPC metadata service")
	}()

	if err := srv.Serve(lis); err != nil {
		logger.Fatal("Failed to serve grpc", zap.Error(err))
	}
	wg.Wait()
}
