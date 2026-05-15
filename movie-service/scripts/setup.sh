#!/bin/bash
set -e
cd "$(dirname "$0")/.."

docker compose down 2>/dev/null || true
docker rm -f cinema-postgres cinema-redis cinema-nats postgres 2>/dev/null || true

docker compose up -d
echo "Waiting for PostgreSQL..."
for i in $(seq 1 30); do
  if docker exec cinema-postgres pg_isready -U postgres -d movie_db >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

go mod tidy
echo "Setup done. Run: go run cmd/server/main.go"
