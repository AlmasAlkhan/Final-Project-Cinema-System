CREATE TABLE IF NOT EXISTS halls (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    rows INT NOT NULL,
    cols INT NOT NULL
);

INSERT INTO halls (name, rows, cols)
SELECT 'Hall 1', 8, 10 WHERE NOT EXISTS (SELECT 1 FROM halls WHERE name = 'Hall 1');
INSERT INTO halls (name, rows, cols)
SELECT 'Hall 2', 6, 8 WHERE NOT EXISTS (SELECT 1 FROM halls WHERE name = 'Hall 2');
