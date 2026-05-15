#!/bin/bash
set -e

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
CINEMA_ROOT="$(cd "$ROOT/.." && pwd)"

cd "$CINEMA_ROOT"
docker compose down 2>/dev/null || true
docker rm -f cinema-postgres cinema-redis cinema-nats 2>/dev/null || true
docker compose up -d

echo "waiting for postgres..."
for i in $(seq 1 40); do
  if docker exec cinema-postgres pg_isready -U user -d movie_db >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

cd "$ROOT"
go mod tidy

if command -v protoc >/dev/null 2>&1; then
  export PATH="$PATH:$(go env GOPATH)/bin"
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  protoc --go_out=proto/movie --go_opt=paths=source_relative \
    --go-grpc_out=proto/movie --go-grpc_opt=paths=source_relative \
    proto/movie.proto
  echo "proto generated"
else
  echo "protoc not found: using committed proto/movie/*.go if present"
fi

go build -o /tmp/movie-grpc ./cmd/server/main.go
echo "setup ok. run: cd $ROOT && go run cmd/server/main.go"
