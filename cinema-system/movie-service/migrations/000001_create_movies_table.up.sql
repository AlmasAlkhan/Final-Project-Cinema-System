CREATE TABLE IF NOT EXISTS movies (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    duration_min INT NOT NULL,
    poster_url TEXT,
    release_date DATE,
    created_at TIMESTAMP DEFAULT NOW()
);
