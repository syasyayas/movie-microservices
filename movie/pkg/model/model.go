package model

import "moviedata.com/metadata/pkg/model"

// MovieDetails includes movie metadata and its aggregated rating.
type MovieDetails struct {
	Rating   *float64       `json:"rating,omitEmpty"`
	Metadata model.Metadata `json:"metadata"`
}
