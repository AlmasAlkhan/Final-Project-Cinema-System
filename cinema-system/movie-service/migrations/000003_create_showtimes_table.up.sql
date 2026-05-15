CREATE TABLE IF NOT EXISTS showtimes (
    id BIGSERIAL PRIMARY KEY,
    movie_id BIGINT REFERENCES movies(id) ON DELETE CASCADE,
    hall_id BIGINT REFERENCES halls(id) ON DELETE CASCADE,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    price DECIMAL(10,2) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_showtimes_movie_id ON showtimes(movie_id);
CREATE INDEX IF NOT EXISTS idx_showtimes_hall_id ON showtimes(hall_id);
