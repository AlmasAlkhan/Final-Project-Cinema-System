package usecase

import (
	"context"
	"time"

	"github.com/yourteam/cinema-system/movie-service/internal/domain"
	"github.com/yourteam/cinema-system/movie-service/internal/repository"
	"github.com/yourteam/cinema-system/movie-service/pkg/natsbus"
)

type ShowtimeUsecase interface {
	CreateShowtime(ctx context.Context, movieID, hallID int64, startTimeStr string, durationMin int32, price float64) (*domain.Showtime, error)
	GetShowtime(ctx context.Context, id int64) (*domain.Showtime, error)
	DeleteShowtime(ctx context.Context, id int64) error
	ListShowtimesByMovie(ctx context.Context, movieID int64) ([]domain.Showtime, error)
	GetHallLayout(ctx context.Context, hallID int64) (*domain.Hall, error)
}

type showtimeUsecase struct {
	showRepo repository.ShowtimeRepository
	hallRepo repository.HallRepository
	bus      *natsbus.Publisher
}

func NewShowtimeUsecase(showRepo repository.ShowtimeRepository, hallRepo repository.HallRepository, bus *natsbus.Publisher) ShowtimeUsecase {
	return &showtimeUsecase{showRepo: showRepo, hallRepo: hallRepo, bus: bus}
}

func (u *showtimeUsecase) CreateShowtime(ctx context.Context, movieID, hallID int64, startTimeStr string, durationMin int32, price float64) (*domain.Showtime, error) {
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		startTime, err = time.Parse("2006-01-02T15:04:05", startTimeStr)
		if err != nil {
			return nil, err
		}
	}
	if durationMin <= 0 {
		durationMin = 90
	}
	endTime := startTime.Add(time.Duration(durationMin) * time.Minute)
	st := &domain.Showtime{
		MovieID:   movieID,
		HallID:    hallID,
		StartTime: startTime,
		EndTime:   endTime,
		Price:     price,
	}
	if err := u.showRepo.Create(ctx, st); err != nil {
		return nil, err
	}
	_ = u.bus.Publish("showtime.created", map[string]any{
		"id":         st.ID,
		"movie_id":   st.MovieID,
		"hall_id":    st.HallID,
		"start_time": st.StartTime.Format(time.RFC3339),
		"end_time":   st.EndTime.Format(time.RFC3339),
		"price":      st.Price,
	})
	return st, nil
}

func (u *showtimeUsecase) GetShowtime(ctx context.Context, id int64) (*domain.Showtime, error) {
	return u.showRepo.GetByID(ctx, id)
}

func (u *showtimeUsecase) DeleteShowtime(ctx context.Context, id int64) error {
	return u.showRepo.Delete(ctx, id)
}

func (u *showtimeUsecase) ListShowtimesByMovie(ctx context.Context, movieID int64) ([]domain.Showtime, error) {
	return u.showRepo.ListByMovie(ctx, movieID)
}

func (u *showtimeUsecase) GetHallLayout(ctx context.Context, hallID int64) (*domain.Hall, error) {
	return u.hallRepo.GetByID(ctx, hallID)
}
