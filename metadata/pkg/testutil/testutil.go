package testutil

import (
	"moviedata.com/gen"
	"moviedata.com/metadata/internal/controller/metadata"
	grpcHandler "moviedata.com/metadata/internal/handler/grpc"
	"moviedata.com/metadata/internal/repository/memory"
)

func NewMetadaGRPCServer() gen.MetadataServiceServer {
	r := memory.New()
	ctrl := metadata.New(r)
	return grpcHandler.New(ctrl)
}
