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

	"github.com/grpc-ecosystem/go-grpc-middleware/ratelimit"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
	"moviedata.com/gen"
	"moviedata.com/movie/internal/controller/movie"
	metadatagateway "moviedata.com/movie/internal/gateway/metadata/grpc"
	ratinggateway "moviedata.com/movie/internal/gateway/rating/grpc"
	grpchandler "moviedata.com/movie/internal/handler/grpc"
	"moviedata.com/pkg/discovery"
	"moviedata.com/pkg/discovery/consul"
)

const serviceName = "movie"
const limit = 100
const burst = 100

func main() {
	logger, _ := zap.NewProduction()
	logger = logger.With(zap.String("service", serviceName))
	f, err := os.Open("base.yaml")
	if err != nil {
		log.Fatal("Failed to open config file", zap.Error(err))
	}
	defer f.Close()

	var cfg serviceConfig

	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		logger.Fatal("Failed to unmarshal config", zap.Error(err))
	}

	logger.Info("Starting the movie service", zap.Int("post", cfg.APIConfig.Port))

	registry, err := consul.NewRegistry("consul:8500")
	if err != nil {
		logger.Fatal("Failed to connect to consul", zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())

	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("movie:%d", cfg.APIConfig.Port)); err != nil {
		logger.Fatal("Failed to register service in consul", zap.Error(err))
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

	metadataGateway := metadatagateway.New(registry)
	ratingGateway := ratinggateway.New(registry)
	ctrl := movie.New(ratingGateway, metadataGateway)
	h := grpchandler.New(ctrl)
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.APIConfig.Port))
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}
	limiter := newLimiter(limit, burst)
	srv := grpc.NewServer(grpc.UnaryInterceptor(ratelimit.UnaryServerInterceptor(limiter)))
	reflection.Register(srv)
	gen.RegisterMovieServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		logger.Fatal("Failed to serve grpc", zap.Error(err))
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s := <-sigChan
		logger.Info("Recived signall %v, attempting graceful shutdown", zap.String("signal", s.String()))
		cancel()
		srv.GracefulStop()
		logger.Info("Gracefully shutted down movie service")
	}()
	wg.Wait()
}

type limiter struct {
	l *rate.Limiter
}

func newLimiter(limit, burst int) *limiter {
	return &limiter{rate.NewLimiter(rate.Limit(limit), burst)}
}

func (l *limiter) Limit() bool {
	return l.l.Allow()
}
