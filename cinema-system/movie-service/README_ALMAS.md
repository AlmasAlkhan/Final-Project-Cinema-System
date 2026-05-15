# Movie Service — часть Алмаса (Clean Architecture + gRPC)

## Структура

```
cinema-system/
  docker-compose.yml          # Postgres :5433, Redis, NATS
  movie-service/
    cmd/server/main.go
    internal/{domain,repository,usecase,delivery/grpc}
    pkg/{config,db,cache,logger,natsbus}
    migrations/
    proto/movie.proto
    scripts/setup.sh
    scripts/test-grpc.sh
```

## Запуск (копируй по одной строке, без строк с #)

```bash
mkdir -p ~/cinema-system
```

Скопируй папку `cinema-system` из проекта в `~/cinema-system` или работай прямо из репозитория.

```bash
cd ~/cinema-system
chmod +x movie-service/scripts/setup.sh movie-service/scripts/test-grpc.sh
./movie-service/scripts/setup.sh
```

В другом терминале:

```bash
cd ~/cinema-system/movie-service
go run cmd/server/main.go
```

В третьем терминале:

```bash
cd ~/cinema-system/movie-service
./scripts/test-grpc.sh
```

## Реализовано

| Требование | Статус |
|------------|--------|
| Clean Architecture (domain, repo, usecase, delivery) | ✅ |
| gRPC: Create/Get/Update/Delete/List Movie | ✅ |
| gRPC: Create/Get/Delete/List Showtimes | ✅ |
| gRPC: GetHallLayout, ListHalls, Health | ✅ |
| PostgreSQL + миграции | ✅ |
| Redis кэш GetMovie (`cache_hit` в ответе) | ✅ |
| NATS `showtime.created` при создании сеанса | ✅ |
| Docker Compose | ✅ (порт Postgres **5433**) |

## gRPC порт

`localhost:50051`

## Пример grpcurl

```bash
grpcurl -plaintext -d '{"title":"Inception","duration_min":148,"release_date":"2010-07-16"}' \
  localhost:50051 movie.MovieService/CreateMovie
```

## Важно

- Не вставляй в терминал строки, начинающиеся с `#` — zsh выдаст `command not found: #`.
- Если порт 5432 занят — в `.env` уже указан **5433**.
- Модуль: `github.com/yourteam/cinema-system/movie-service` — локально работает без GitHub.
