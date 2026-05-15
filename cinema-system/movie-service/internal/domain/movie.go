package domain

import "time"

type Movie struct {
	ID          int64
	Title       string
	Description string
	DurationMin int32
	PosterURL   string
	ReleaseDate time.Time
	CreatedAt   time.Time
}

type Hall struct {
	ID   int64
	Name string
	Rows int32
	Cols int32
}

type Showtime struct {
	ID        int64
	MovieID   int64
	HallID    int64
	StartTime time.Time
	EndTime   time.Time
	Price     float64
}
