package model

import "time"

type CarImage struct {
	ID        int64
	CarID     int64
	ImageURL  string
	AltText   string
	SortOrder int
	IsPrimary bool
	CreatedAt time.Time
}
