package grpc

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/yourteam/cinema-system/movie-service/internal/domain"
	"github.com/yourteam/cinema-system/movie-service/internal/repository"
	"github.com/yourteam/cinema-system/movie-service/internal/usecase"
	pb "github.com/yourteam/cinema-system/movie-service/proto/movie"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MovieGRPCServer struct {
	pb.UnimplementedMovieServiceServer
	movieUC    usecase.MovieUsecase
	showtimeUC usecase.ShowtimeUsecase
	hallRepo   repository.HallRepository
}

func NewMovieGRPCServer(movieUC usecase.MovieUsecase, showtimeUC usecase.ShowtimeUsecase, hallRepo repository.HallRepository) *MovieGRPCServer {
	return &MovieGRPCServer{movieUC: movieUC, showtimeUC: showtimeUC, hallRepo: hallRepo}
}

func (s *MovieGRPCServer) Health(ctx context.Context, _ *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{Status: "ok"}, nil
}

func (s *MovieGRPCServer) CreateMovie(ctx context.Context, req *pb.CreateMovieRequest) (*pb.MovieResponse, error) {
	movie, err := s.movieUC.CreateMovie(ctx, req.Title, req.Description, req.DurationMin, req.PosterUrl, req.ReleaseDate)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toMoviePB(movie, false), nil
}

func (s *MovieGRPCServer) GetMovie(ctx context.Context, req *pb.GetMovieRequest) (*pb.MovieResponse, error) {
	movie, cacheHit, err := s.movieUC.GetMovie(ctx, req.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "movie not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toMoviePB(movie, cacheHit), nil
}

func (s *MovieGRPCServer) UpdateMovie(ctx context.Context, req *pb.UpdateMovieRequest) (*pb.MovieResponse, error) {
	movie, err := s.movieUC.UpdateMovie(ctx, req.Id, req.Title, req.Description, req.DurationMin, req.PosterUrl, req.ReleaseDate)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toMoviePB(movie, false), nil
}

func (s *MovieGRPCServer) DeleteMovie(ctx context.Context, req *pb.DeleteMovieRequest) (*pb.Empty, error) {
	if err := s.movieUC.DeleteMovie(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Empty{}, nil
}

func (s *MovieGRPCServer) ListMovies(ctx context.Context, req *pb.ListMoviesRequest) (*pb.ListMoviesResponse, error) {
	movies, total, err := s.movieUC.ListMovies(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := &pb.ListMoviesResponse{Total: total}
	for i := range movies {
		resp.Movies = append(resp.Movies, toMoviePB(&movies[i], false))
	}
	return resp, nil
}

func (s *MovieGRPCServer) CreateShowtime(ctx context.Context, req *pb.CreateShowtimeRequest) (*pb.ShowtimeResponse, error) {
	st, err := s.showtimeUC.CreateShowtime(ctx, req.MovieId, req.HallId, req.StartTime, req.DurationMin, req.Price)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toShowtimePB(st), nil
}

func (s *MovieGRPCServer) GetShowtime(ctx context.Context, req *pb.GetShowtimeRequest) (*pb.ShowtimeResponse, error) {
	st, err := s.showtimeUC.GetShowtime(ctx, req.Id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "showtime not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toShowtimePB(st), nil
}

func (s *MovieGRPCServer) DeleteShowtime(ctx context.Context, req *pb.DeleteShowtimeRequest) (*pb.Empty, error) {
	if err := s.showtimeUC.DeleteShowtime(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.Empty{}, nil
}

func (s *MovieGRPCServer) ListShowtimes(ctx context.Context, req *pb.ListShowtimesRequest) (*pb.ListShowtimesResponse, error) {
	items, err := s.showtimeUC.ListShowtimesByMovie(ctx, req.MovieId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := &pb.ListShowtimesResponse{}
	for i := range items {
		resp.Showtimes = append(resp.Showtimes, toShowtimePB(&items[i]))
	}
	return resp, nil
}

func (s *MovieGRPCServer) GetHallLayout(ctx context.Context, req *pb.GetHallLayoutRequest) (*pb.HallLayoutResponse, error) {
	h, err := s.showtimeUC.GetHallLayout(ctx, req.HallId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "hall not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toHallPB(h), nil
}

func (s *MovieGRPCServer) ListHalls(ctx context.Context, _ *pb.ListHallsRequest) (*pb.ListHallsResponse, error) {
	halls, err := s.hallRepo.List(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := &pb.ListHallsResponse{}
	for i := range halls {
		resp.Halls = append(resp.Halls, toHallPB(&halls[i]))
	}
	return resp, nil
}

func toMoviePB(movie *domain.Movie, cacheHit bool) *pb.MovieResponse {
	if movie == nil {
		return nil
	}
	created := ""
	if !movie.CreatedAt.IsZero() {
		created = movie.CreatedAt.UTC().Format(time.RFC3339)
	}
	return &pb.MovieResponse{
		Id:          movie.ID,
		Title:       movie.Title,
		Description: movie.Description,
		DurationMin: movie.DurationMin,
		PosterUrl:   movie.PosterURL,
		ReleaseDate: movie.ReleaseDate.Format("2006-01-02"),
		CreatedAt:   created,
		CacheHit:    cacheHit,
	}
}

func toShowtimePB(st *domain.Showtime) *pb.ShowtimeResponse {
	return &pb.ShowtimeResponse{
		Id:        st.ID,
		MovieId:   st.MovieID,
		HallId:    st.HallID,
		StartTime: st.StartTime.UTC().Format(time.RFC3339),
		EndTime:   st.EndTime.UTC().Format(time.RFC3339),
		Price:     st.Price,
	}
}

func toHallPB(h *domain.Hall) *pb.HallLayoutResponse {
	return &pb.HallLayoutResponse{
		Id:   h.ID,
		Name: h.Name,
		Rows: h.Rows,
		Cols: h.Cols,
	}
}
