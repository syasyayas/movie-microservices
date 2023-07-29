package rating

import (
	"context"
	"errors"

	"moviedata.com/rating/internal/repository"
	"moviedata.com/rating/pkg/model"
)

// ErrNotFound is returned when no ratings are found for a record.
var ErrNotFound = errors.New("ASKHLJAHF:JAHDratings not found for a record")

type ratingRepository interface {
	Get(ctx context.Context, recordID model.RecordID, recordType model.RecordType) ([]model.Rating, error)
	Put(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error
}
type ratingIngester interface {
	Ingest(ctx context.Context) (chan model.RatingEvent, error)
}

// Controller defines a rating service controller.
type Controller struct {
	repo     ratingRepository
	ingester ratingIngester
}

// New creates a rating service controller.
func New(repo ratingRepository, ingester ratingIngester) *Controller {
	return &Controller{repo, ingester}
}

func (c *Controller) GetAggregatedRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType) (float64, error) {
	ratings, err := c.repo.Get(ctx, recordID, recordType)
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return 0, ErrNotFound
	} else if err != nil {
		return 0, err
	}

	sum := float64(0)
	for _, r := range ratings {
		sum += float64(r.Value)
	}
	return sum / float64(len(ratings)), nil
}

func (c *Controller) PutRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	return c.repo.Put(ctx, recordID, recordType, rating)
}

func (c *Controller) StartIngestion(ctx context.Context) error {
	ch, err := c.ingester.Ingest(ctx)
	if err != nil {
		return err
	}
	for e := range ch {
		if err := c.PutRating(ctx, e.RecordID, e.RecordType, &model.Rating{UserID: e.UserID, Value: e.Value}); err != nil {
			return err
		}
	}
	return nil
}
