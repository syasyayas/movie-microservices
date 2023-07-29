package mysql

import (
	"context"
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"moviedata.com/rating/internal/repository"
	"moviedata.com/rating/pkg/model"
)

type Repository struct {
	db *sql.DB
}

func New() (*Repository, error) {
	db, err := sql.Open("mysql", "root:password@tcp(mysql:3306)/movieexample")
	if err != nil {
		return nil, err
	}

	return &Repository{db}, nil
}

func (r *Repository) Get(ctx context.Context, recordID model.RecordID, recordType model.RecordType) ([]model.Rating, error) {
	row, err := r.db.QueryContext(ctx, "SELECT user_id, value FROM ratings WHERE record_id = ? AND record_type = ?", recordID, recordType)
	if err != nil {
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
	return res, nil
}

func (r *Repository) Put(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO ratings (record_id, record_type, user_id, value) VALUES (?,?,?,?)",
		recordID,
		recordType,
		rating.UserID,
		rating.Value)
	return err
}
