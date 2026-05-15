package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/redis/go-redis/v9"
	_ "github.com/lib/pq"
)

type Movie struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	DurationMin int    `json:"duration_min"`
	PosterURL   string `json:"poster_url,omitempty"`
	ReleaseDate string `json:"release_date"`
	CreatedAt   string `json:"created_at,omitempty"`
}

var (
	db          *sql.DB
	redisClient *redis.Client
	natsConn    *nats.Conn
)

func main() {
	connStr := getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5433/movie_db?sslmode=disable")
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("PostgreSQL: %v", err)
	}
	if err = migrate(db); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	redisClient = redis.NewClient(&redis.Options{Addr: getenv("REDIS_ADDR", "localhost:6379")})
	if err = redisClient.Ping(context.Background()).Err(); err != nil {
		log.Printf("Redis not available: %v", err)
		redisClient = nil
	} else {
		log.Println("Connected to Redis")
	}

	natsConn, err = nats.Connect(getenv("NATS_URL", nats.DefaultURL))
	if err != nil {
		log.Printf("NATS not available: %v", err)
	} else {
		log.Println("Connected to NATS")
		defer natsConn.Close()
	}

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/movies", moviesHandler)
	http.HandleFunc("/movies/", movieByIDHandler)

	log.Println("HTTP server: http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS movies (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    duration_min INT NOT NULL,
    poster_url TEXT,
    release_date TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
)`)
	return err
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func moviesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		listMovies(w)
	case http.MethodPost:
		createMovie(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func listMovies(w http.ResponseWriter) {
	rows, err := db.Query(`SELECT id, title, COALESCE(description,''), duration_min, COALESCE(poster_url,''), release_date, created_at FROM movies ORDER BY id`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var movies []Movie
	for rows.Next() {
		var m Movie
		var created time.Time
		if err := rows.Scan(&m.ID, &m.Title, &m.Description, &m.DurationMin, &m.PosterURL, &m.ReleaseDate, &created); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		m.CreatedAt = created.UTC().Format(time.RFC3339)
		movies = append(movies, m)
	}
	if movies == nil {
		movies = []Movie{}
	}
	_ = json.NewEncoder(w).Encode(movies)
}

func createMovie(w http.ResponseWriter, r *http.Request) {
	var m Movie
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if m.Title == "" || m.DurationMin <= 0 || m.ReleaseDate == "" {
		http.Error(w, "title, duration_min, release_date required", http.StatusBadRequest)
		return
	}

	var created time.Time
	err := db.QueryRow(
		`INSERT INTO movies (title, description, duration_min, poster_url, release_date) VALUES ($1,$2,$3,$4,$5) RETURNING id, created_at`,
		m.Title, m.Description, m.DurationMin, m.PosterURL, m.ReleaseDate,
	).Scan(&m.ID, &created)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	m.CreatedAt = created.UTC().Format(time.RFC3339)

	if redisClient != nil {
		_ = redisClient.Del(context.Background(), cacheKey(m.ID))
	}
	publishEvent("movie.created", m)

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(m)
}

func movieByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 2 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(parts[1])
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	key := cacheKey(id)
	if redisClient != nil {
		if cached, err := redisClient.Get(context.Background(), key).Bytes(); err == nil {
			w.Header().Set("X-Cache", "HIT")
			w.Write(cached)
			return
		}
	}

	var m Movie
	var created time.Time
	err = db.QueryRow(
		`SELECT id, title, COALESCE(description,''), duration_min, COALESCE(poster_url,''), release_date, created_at FROM movies WHERE id=$1`,
		id,
	).Scan(&m.ID, &m.Title, &m.Description, &m.DurationMin, &m.PosterURL, &m.ReleaseDate, &created)
	if err == sql.ErrNoRows {
		http.Error(w, "movie not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	m.CreatedAt = created.UTC().Format(time.RFC3339)

	body, _ := json.Marshal(m)
	if redisClient != nil {
		_ = redisClient.Set(context.Background(), key, body, 5*time.Minute).Err()
	}
	w.Header().Set("X-Cache", "MISS")
	w.Write(body)
}

func cacheKey(id int) string {
	return "movie:" + strconv.Itoa(id)
}

func publishEvent(subject string, payload any) {
	if natsConn == nil {
		return
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	if err := natsConn.Publish(subject, data); err != nil {
		log.Printf("nats publish %s: %v", subject, err)
		return
	}
	log.Printf("Published %s", subject)
}
