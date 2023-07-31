package grpc

import (
	"context"
	"errors"
	"log"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"moviedata.com/gen"
	"moviedata.com/rating/internal/controller/rating"
	"moviedata.com/rating/pkg/model"
)

type Handler struct {
	gen.UnsafeRatingServiceServer
	ctrl   *rating.Controller
	logger *zap.Logger
}

func New(ctrl *rating.Controller, logger *zap.Logger) *Handler {
	logger = logger.With(zap.String("component", "grpchandler"))
	return &Handler{ctrl: ctrl, logger: logger}
}

func (h *Handler) GetAggregatedRating(ctx context.Context, req *gen.GetAggregatedRatingRequest) (*gen.GetAggregatedRatingResponse, error) {
	if req == nil || req.RecordId == "" || req.RecordType == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil req or empty")
	}
	v, err := h.ctrl.GetAggregatedRating(ctx, model.RecordID(req.RecordId), model.RecordType(req.RecordType))
	if err != nil && errors.Is(err, rating.ErrNotFound) {
		return nil, status.Errorf(codes.NotFound, err.Error())
	} else if err != nil {
		h.logger.Error("Failed to process request", zap.Error(err), zap.String("endpoint", "GetAggregatedRating"))
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &gen.GetAggregatedRatingResponse{RatingValue: v}, nil
}

func (h *Handler) PutRating(ctx context.Context, req *gen.PutRatingRequest) (*gen.PutRatingResponse, error) {
	log.Printf("Handling put request")
	if req == nil || req.RecordId == "" || req.UserId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "nil req or empty UserId or RecordId")
	}
	err := h.ctrl.PutRating(ctx,
		model.RecordID(req.RecordId),
		model.RecordType(req.RecordType),
		&model.Rating{
			UserID: model.UserID(req.UserId),
			Value:  model.RatingValue(req.RatingValue),
		})
	if err != nil {
		h.logger.Error("Failed to process request", zap.Error(err), zap.String("endpoint", "PutRating"))
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &gen.PutRatingResponse{}, nil
}
