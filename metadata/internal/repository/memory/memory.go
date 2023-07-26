package memory

import (
	"context"
	"sync"

	"moviedata.com/metadata/internal/repository"
	"moviedata.com/metadata/pkg/model"
)

type Repository struct {
	sync.RWMutex
	data map[string]*model.Metadata
}

func New() *Repository {
	return &Repository{data: map[string]*model.Metadata{}}
}

func (r *Repository) Get(_ context.Context, id string) (model *model.Metadata, err error) {
	r.Lock()
	defer r.Unlock()
	model, ok := r.data[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	return model, nil
}

func (r *Repository) Put(_ context.Context, id string, metadata *model.Metadata) error {
	r.Lock()
	defer r.Unlock()
	r.data[id] = metadata
	return nil
}
