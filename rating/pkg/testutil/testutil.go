package testutil

import (
	"context"

	"go.uber.org/zap"
	"moviedata.com/gen"
	"moviedata.com/rating/internal/controller/rating"
	grpcHandler "moviedata.com/rating/internal/handler/grpc"
	"moviedata.com/rating/internal/repository/memory"
	"moviedata.com/rating/pkg/model"
)

func NewTestRatingGRPCServer() gen.RatingServiceServer {
	r := memory.New()
	logger, _ := zap.NewProduction()
	r.Put(context.Background(), "nil", "movie", &model.Rating{})
	ctrl := rating.New(r, nil, logger)
	return grpcHandler.New(ctrl, logger)
}
