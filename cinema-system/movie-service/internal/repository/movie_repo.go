package repository

import (
	"context"
	"database/sql"

	"github.com/yourteam/cinema-system/movie-service/internal/domain"
)

type MovieRepository interface {
	Create(ctx context.Context, m *domain.Movie) error
	GetByID(ctx context.Context, id int64) (*domain.Movie, error)
	Update(ctx context.Context, m *domain.Movie) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, page, pageSize int32) ([]domain.Movie, int32, error)
}

type movieRepo struct {
	db *sql.DB
}

func NewMovieRepository(db *sql.DB) MovieRepository {
	return &movieRepo{db: db}
}

func (r *movieRepo) Create(ctx context.Context, m *domain.Movie) error {
	query := `INSERT INTO movies (title, description, duration_min, poster_url, release_date)
              VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, m.Title, m.Description, m.DurationMin, m.PosterURL, m.ReleaseDate).
		Scan(&m.ID, &m.CreatedAt)
}

func (r *movieRepo) GetByID(ctx context.Context, id int64) (*domain.Movie, error) {
	query := `SELECT id, title, description, duration_min, poster_url, release_date, created_at
              FROM movies WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	m := &domain.Movie{}
	err := row.Scan(&m.ID, &m.Title, &m.Description, &m.DurationMin, &m.PosterURL, &m.ReleaseDate, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (r *movieRepo) Update(ctx context.Context, m *domain.Movie) error {
	query := `UPDATE movies SET title=$1, description=$2, duration_min=$3, poster_url=$4, release_date=$5 WHERE id=$6`
	_, err := r.db.ExecContext(ctx, query, m.Title, m.Description, m.DurationMin, m.PosterURL, m.ReleaseDate, m.ID)
	return err
}

func (r *movieRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM movies WHERE id=$1`, id)
	return err
}

func (r *movieRepo) List(ctx context.Context, page, pageSize int32) ([]domain.Movie, int32, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	rows, err := r.db.QueryContext(ctx, `SELECT id, title, description, duration_min, poster_url, release_date, created_at
                                         FROM movies ORDER BY id LIMIT $1 OFFSET $2`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var movies []domain.Movie
	for rows.Next() {
		var m domain.Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.DurationMin, &m.PosterURL, &m.ReleaseDate, &m.CreatedAt); err != nil {
			return nil, 0, err
		}
		movies = append(movies, m)
	}
	var total int32
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM movies`).Scan(&total)
	if movies == nil {
		movies = []domain.Movie{}
	}
	return movies, total, nil
}
