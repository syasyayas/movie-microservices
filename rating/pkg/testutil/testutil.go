package testutil

import (
	"context"

	"moviedata.com/gen"
	"moviedata.com/rating/internal/controller/rating"
	grpcHandler "moviedata.com/rating/internal/handler/grpc"
	"moviedata.com/rating/internal/repository/memory"
	"moviedata.com/rating/pkg/model"
)

func NewTestRatingGRPCServer() gen.RatingServiceServer {
	r := memory.New()
	r.Put(context.Background(), "nil", "movie", &model.Rating{})
	ctrl := rating.New(r, nil)
	return grpcHandler.New(ctrl)
}
