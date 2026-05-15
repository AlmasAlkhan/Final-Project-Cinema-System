package main

import (
	"context"
	"log"
	"net"
	"path/filepath"

	grpcdelivery "github.com/yourteam/cinema-system/movie-service/internal/delivery/grpc"
	"github.com/yourteam/cinema-system/movie-service/internal/repository"
	"github.com/yourteam/cinema-system/movie-service/internal/usecase"
	"github.com/yourteam/cinema-system/movie-service/pkg/cache"
	"github.com/yourteam/cinema-system/movie-service/pkg/config"
	"github.com/yourteam/cinema-system/movie-service/pkg/db"
	"github.com/yourteam/cinema-system/movie-service/pkg/logger"
	"github.com/yourteam/cinema-system/movie-service/pkg/natsbus"
	pb "github.com/yourteam/cinema-system/movie-service/proto/movie"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := logger.Init(); err != nil {
		log.Fatalf("logger: %v", err)
	}
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	sqlDB, err := db.NewPostgresConnection(cfg.PostgresURL)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}
	defer sqlDB.Close()

	migrationsDir := filepath.Join("migrations")
	if err := db.RunMigrations(sqlDB, migrationsDir); err != nil {
		log.Fatalf("migrations: %v", err)
	}
	log.Println("migrations applied")

	redisClient := cache.NewRedisClient(cfg.RedisAddr)
	defer redisClient.Close()
	redisCache := cache.NewRedisCache(redisClient)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Printf("redis warning: %v", err)
		redisCache = nil
	} else {
		log.Println("connected to redis")
	}

	bus, err := natsbus.NewPublisher(cfg.NATSURL)
	if err != nil {
		log.Printf("nats warning: %v", err)
	} else {
		defer bus.Close()
		log.Println("connected to nats")
	}

	movieRepo := repository.NewMovieRepository(sqlDB)
	showtimeRepo := repository.NewShowtimeRepository(sqlDB)
	hallRepo := repository.NewHallRepository(sqlDB)

	movieUC := usecase.NewMovieUsecase(movieRepo, redisCache)
	showtimeUC := usecase.NewShowtimeUsecase(showtimeRepo, hallRepo, bus)

	grpcServer := grpc.NewServer()
	movieServer := grpcdelivery.NewMovieGRPCServer(movieUC, showtimeUC, hallRepo)
	pb.RegisterMovieServiceServer(grpcServer, movieServer)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	log.Printf("Movie Service gRPC on :%s", cfg.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
