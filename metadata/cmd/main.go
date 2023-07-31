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
	f, err := os.Open("base.yaml")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var cfg serviceConfig

	registry, err := consul.NewRegistry("consul:8500")
	if err != nil {
		panic(err)
	}

	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		panic(err)
	}
	log.Printf("Starting the movie metadata service on port %d", cfg.APIConfig.Port)

	ctx, cancel := context.WithCancel(context.Background())

	instanceID := discovery.GenerateInstanceID(serviceName)

	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("metadata:%d", cfg.APIConfig.Port)); err != nil {
		panic(err)
	}

	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				log.Println("Failed to report healthy state: " + err.Error())
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
		log.Fatalf("failed to listen: %v", err)
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
		log.Printf("Received signall %v, attemting graceful shutdown", s)
		srv.GracefulStop()
		log.Println("Gracefully stopped gRPC metadata service")
	}()

	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
	wg.Wait()
}
