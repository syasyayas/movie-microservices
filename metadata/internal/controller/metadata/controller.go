package metadata

import (
	"context"
	"errors"

	"moviedata.com/metadata/internal/repository"
	"moviedata.com/metadata/pkg/model"
)

var ErrNotFound = errors.New("not found")

type metadataRepository interface {
	Get(ctx context.Context, id string) (*model.Metadata, error)
	Put(ctx context.Context, metadataID string, metadata *model.Metadata) error
}

type Controller struct {
	repo metadataRepository
}

func New(repo metadataRepository) *Controller {
	return &Controller{repo}
}

func (c *Controller) Get(ctx context.Context, id string) (model *model.Metadata, err error) {
	model, err = c.repo.Get(ctx, id)
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	return model, err
}

func (c *Controller) Put(ctx context.Context, metadata *model.Metadata) error {
	return c.repo.Put(ctx, metadata.ID, metadata)
}
