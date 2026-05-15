package repository

import (
	"context"
	"database/sql"

	"github.com/yourteam/cinema-system/movie-service/internal/domain"
)

type ShowtimeRepository interface {
	Create(ctx context.Context, s *domain.Showtime) error
	GetByID(ctx context.Context, id int64) (*domain.Showtime, error)
	Delete(ctx context.Context, id int64) error
	ListByMovie(ctx context.Context, movieID int64) ([]domain.Showtime, error)
}

type HallRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.Hall, error)
	List(ctx context.Context) ([]domain.Hall, error)
}

type showtimeRepo struct {
	db *sql.DB
}

type hallRepo struct {
	db *sql.DB
}

func NewShowtimeRepository(db *sql.DB) ShowtimeRepository {
	return &showtimeRepo{db: db}
}

func NewHallRepository(db *sql.DB) HallRepository {
	return &hallRepo{db: db}
}

func (r *showtimeRepo) Create(ctx context.Context, s *domain.Showtime) error {
	query := `INSERT INTO showtimes (movie_id, hall_id, start_time, end_time, price)
              VALUES ($1, $2, $3, $4, $5) RETURNING id`
	return r.db.QueryRowContext(ctx, query, s.MovieID, s.HallID, s.StartTime, s.EndTime, s.Price).Scan(&s.ID)
}

func (r *showtimeRepo) GetByID(ctx context.Context, id int64) (*domain.Showtime, error) {
	query := `SELECT id, movie_id, hall_id, start_time, end_time, price FROM showtimes WHERE id=$1`
	row := r.db.QueryRowContext(ctx, query, id)
	s := &domain.Showtime{}
	err := row.Scan(&s.ID, &s.MovieID, &s.HallID, &s.StartTime, &s.EndTime, &s.Price)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *showtimeRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM showtimes WHERE id=$1`, id)
	return err
}

func (r *showtimeRepo) ListByMovie(ctx context.Context, movieID int64) ([]domain.Showtime, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, movie_id, hall_id, start_time, end_time, price
                                         FROM showtimes WHERE movie_id=$1 ORDER BY start_time`, movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var showtimes []domain.Showtime
	for rows.Next() {
		var s domain.Showtime
		if err := rows.Scan(&s.ID, &s.MovieID, &s.HallID, &s.StartTime, &s.EndTime, &s.Price); err != nil {
			return nil, err
		}
		showtimes = append(showtimes, s)
	}
	if showtimes == nil {
		showtimes = []domain.Showtime{}
	}
	return showtimes, nil
}

func (r *hallRepo) GetByID(ctx context.Context, id int64) (*domain.Hall, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, name, rows, cols FROM halls WHERE id=$1`, id)
	h := &domain.Hall{}
	if err := row.Scan(&h.ID, &h.Name, &h.Rows, &h.Cols); err != nil {
		return nil, err
	}
	return h, nil
}
