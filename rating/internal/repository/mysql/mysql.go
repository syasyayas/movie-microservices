package mysql

import (
	"context"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
	"moviedata.com/rating/internal/repository"
	"moviedata.com/rating/pkg/model"
)

type Repository struct {
	logger *zap.Logger
	db     *sql.DB
}

func New(logger *zap.Logger) (*Repository, error) {
	db, err := sql.Open("mysql", "root:password@tcp(mysql:3306)/movieexample")
	if err != nil {
		return nil, err
	}

	return &Repository{db: db, logger: logger.With(zap.String("component", "repository"))}, nil
}

func (r *Repository) Get(ctx context.Context, recordID model.RecordID, recordType model.RecordType) ([]model.Rating, error) {
	r.logger.Debug("Executing query", zap.String("method", "Get"), zap.String("record_id", string(recordID)), zap.String("record_type", string(recordType)))
	row, err := r.db.QueryContext(ctx, "SELECT user_id, value FROM ratings WHERE record_id = ? AND record_type = ?", recordID, recordType)
	if err != nil {
		r.logger.Error("Failed executing query", zap.String("method", "Get"), zap.String("record_id", string(recordID)), zap.String("record_type", string(recordType)), zap.Error(err))
		return nil, err
	}
	defer row.Close()
	var res []model.Rating
	for row.Next() {
		var userID string
		var value int32
		if err := row.Scan(&userID, &value); err != nil {
			return nil, err
		}
		res = append(res, model.Rating{UserID: model.UserID(userID), Value: model.RatingValue(value)})
	}
	if len(res) == 0 {
		return nil, repository.ErrNotFound
	}
	r.logger.Debug("Successfully executed query")
	return res, nil
}

func (r *Repository) Put(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	r.logger.Debug("Executing put method")
	_, err := r.db.ExecContext(ctx, "INSERT INTO ratings (record_id, record_type, user_id, value) VALUES (?,?,?,?)",
		recordID,
		recordType,
		rating.UserID,
		rating.Value)
	return err
}
