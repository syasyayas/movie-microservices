package metadata

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	gen "moviedata.com/gen/mock/metadata/repository"
	"moviedata.com/metadata/internal/repository"
	"moviedata.com/metadata/pkg/model"
)

func TestController(t *testing.T) {
	testCases := []struct {
		desc       string
		expRepoRes *model.Metadata
		expRepoErr error
		wantRes    *model.Metadata
		wantErr    error
	}{
		{
			desc:       "Not found",
			expRepoErr: repository.ErrNotFound,
			wantErr:    ErrNotFound,
		},
		{
			desc:       "unexpected error",
			expRepoErr: errors.New("unexpected error"),
			wantErr:    errors.New("unexpected error"),
		},
		{
			desc:       "success",
			expRepoRes: &model.Metadata{},
			wantRes:    &model.Metadata{},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			repoMock := gen.NewMockmetadataRepository(ctrl)
			c := New(repoMock)
			ctx := context.Background()
			id := "id"
			repoMock.EXPECT().Get(ctx, id).Return(tC.expRepoRes, tC.expRepoErr)
			res, err := c.Get(ctx, id)
			assert.Equal(t, tC.wantRes, res, tC.desc)
			assert.Equal(t, tC.wantErr, err, tC.desc)

		})
	}
}
