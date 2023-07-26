package grpc

import (
	"context"

	"moviedata.com/gen"
	"moviedata.com/metadata/pkg/model"
	"moviedata.com/movie/internal/grpcutil"
	"moviedata.com/pkg/discovery"
)

type Gateway struct {
	registry discovery.Registry
}

func New(reg discovery.Registry) *Gateway {
	return &Gateway{reg}
}

func (g *Gateway) Get(ctx context.Context, id string) (*model.Metadata, error) {
	conn, err := grpcutil.ServiceConnection(ctx, "metadata", g.registry)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := gen.NewMetadataServiceClient(conn)
	resp, err := client.GetMetadata(ctx, &gen.GetMetadataRequest{MovieId: id})
	if err != nil {
		return nil, err
	}
	return model.MetadataFromProto(resp.Metadata), nil
}
