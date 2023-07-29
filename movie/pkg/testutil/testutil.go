package testutil

import (
	"moviedata.com/gen"
	"moviedata.com/movie/internal/controller/movie"
	metadataGateway "moviedata.com/movie/internal/gateway/metadata/grpc"
	ratingGateway "moviedata.com/movie/internal/gateway/rating/grpc"
	grpcHandler "moviedata.com/movie/internal/handler/grpc"
	"moviedata.com/pkg/discovery"
)

func NewTestMovieGRPCServer(registry discovery.Registry) gen.MovieServiceServer {
	metadataGateway := metadataGateway.New(registry)
	ratingGateway := ratingGateway.New(registry)
	ctrl := movie.New(ratingGateway, metadataGateway)
	return grpcHandler.New(ctrl)
}
