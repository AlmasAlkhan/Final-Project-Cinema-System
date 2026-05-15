package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yourteam/cinema-system/movie-service/internal/domain"
	"github.com/yourteam/cinema-system/movie-service/internal/repository"
	"github.com/yourteam/cinema-system/movie-service/pkg/cache"
)

type MovieUsecase interface {
	CreateMovie(ctx context.Context, title, desc string, dur int32, poster, releaseDateStr string) (*domain.Movie, error)
	GetMovie(ctx context.Context, id int64) (*domain.Movie, bool, error)
	UpdateMovie(ctx context.Context, id int64, title, desc string, dur int32, poster, releaseDateStr string) (*domain.Movie, error)
	DeleteMovie(ctx context.Context, id int64) error
	ListMovies(ctx context.Context, page, pageSize int32) ([]domain.Movie, int32, error)
}

type movieUsecase struct {
	repo  repository.MovieRepository
	cache *cache.RedisCache
}

func NewMovieUsecase(repo repository.MovieRepository, c *cache.RedisCache) MovieUsecase {
	return &movieUsecase{repo: repo, cache: c}
}

func movieCacheKey(id int64) string {
	return fmt.Sprintf("movie:%d", id)
}

func (u *movieUsecase) CreateMovie(ctx context.Context, title, desc string, dur int32, poster, releaseDateStr string) (*domain.Movie, error) {
	releaseDate, err := time.Parse("2006-01-02", releaseDateStr)
	if err != nil {
		return nil, err
	}
	movie := &domain.Movie{
		Title:       title,
		Description: desc,
		DurationMin: dur,
		PosterURL:   poster,
		ReleaseDate: releaseDate,
	}
	if err := u.repo.Create(ctx, movie); err != nil {
		return nil, err
	}
	return movie, nil
}

func (u *movieUsecase) GetMovie(ctx context.Context, id int64) (*domain.Movie, bool, error) {
	key := movieCacheKey(id)
	if u.cache != nil {
		if data, err := u.cache.Get(ctx, key); err == nil {
			var m domain.Movie
			if json.Unmarshal(data, &m) == nil {
				return &m, true, nil
			}
		} else if err != redis.Nil {
			return nil, false, err
		}
	}

	movie, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, false, err
	}
	if u.cache != nil {
		if data, err := json.Marshal(movie); err == nil {
			_ = u.cache.Set(ctx, key, data, 5*time.Minute)
		}
	}
	return movie, false, nil
}

func (u *movieUsecase) UpdateMovie(ctx context.Context, id int64, title, desc string, dur int32, poster, releaseDateStr string) (*domain.Movie, error) {
	releaseDate, err := time.Parse("2006-01-02", releaseDateStr)
	if err != nil {
		return nil, err
	}
	movie := &domain.Movie{
		ID:          id,
		Title:       title,
		Description: desc,
		DurationMin: dur,
		PosterURL:   poster,
		ReleaseDate: releaseDate,
	}
	if err := u.repo.Update(ctx, movie); err != nil {
		return nil, err
	}
	if u.cache != nil {
		_ = u.cache.Delete(ctx, movieCacheKey(id))
	}
	return u.repo.GetByID(ctx, id)
}

func (u *movieUsecase) DeleteMovie(ctx context.Context, id int64) error {
	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}
	if u.cache != nil {
		_ = u.cache.Delete(ctx, movieCacheKey(id))
	}
	return nil
}

func (u *movieUsecase) ListMovies(ctx context.Context, page, pageSize int32) ([]domain.Movie, int32, error) {
	return u.repo.List(ctx, page, pageSize)
}
