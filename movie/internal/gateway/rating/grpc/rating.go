package grpc

import (
	"context"

	"moviedata.com/gen"
	"moviedata.com/movie/internal/grpcutil"
	"moviedata.com/pkg/discovery"
	"moviedata.com/rating/pkg/model"
)

type Gateway struct {
	registry discovery.Registry
}

func New(registry discovery.Registry) *Gateway {
	return &Gateway{registry}
}

func (g *Gateway) GetAggregatedRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType) (float64, error) {
	conn, err := grpcutil.ServiceConnection(ctx, "rating", g.registry)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	client := gen.NewRatingServiceClient(conn)
	resp, err := client.GetAggregatedRating(
		ctx,
		&gen.GetAggregatedRatingRequest{RecordId: string(recordID), RecordType: string(recordType)},
	)
	if err != nil {
		return 0, err
	}
	return resp.RatingValue, nil
}

func (g *Gateway) PutRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	return nil
}
