package grpc

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"moviedata.com/gen"
	"moviedata.com/metadata/pkg/model"
	"moviedata.com/movie/internal/controller/movie"
)

type Handler struct {
	gen.UnsafeMovieServiceServer
	ctrl *movie.Controller
}

func New(ctrl *movie.Controller) *Handler {
	return &Handler{ctrl: ctrl}
}

func (h *Handler) GetMovieDetails(ctx context.Context, req *gen.GetMovieDetailsRequest) (*gen.GetMovieDetailsReposne, error) {
	if req == nil || req.MovieId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil req or empty id")
	}

	m, err := h.ctrl.Get(ctx, req.MovieId)
	if err != nil && errors.Is(err, movie.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, err.Error())
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &gen.GetMovieDetailsReposne{
		MovieDetails: &gen.MovieDetails{
			Metadata: model.MetadataToProto(&m.Metadata),
			Rating:   *m.Rating,
		},
	}, nil
}

func (h *Handler) Hello(ctx context.Context, req *gen.HelloRequest) (*gen.HelloResponse, error) {
	if req == nil || req.Name == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil req or empty name")
	}
	return &gen.HelloResponse{
		Hello: fmt.Sprintf("Hello, %s!", req.Name),
	}, nil
}
