# Movie Service — часть Алмаса

## Быстрый старт (копируй по одной строке)

```bash
cd movie-service
chmod +x scripts/setup.sh scripts/test-api.sh
./scripts/setup.sh
```

В **другом** терминале:

```bash
cd movie-service
go run cmd/server/main.go
```

В **третьем** терминале:

```bash
cd movie-service
./scripts/test-api.sh
```

## Важно про терминал

- Не вставляй строки, которые начинаются с `#` — zsh пишет `command not found: #`.
- Не вставляй текст вроде «второй раз HIT» вместе с `curl` — это ломает команду.
- PostgreSQL здесь на порту **5433** (5432 часто занят на Mac).

## Что реализовано

- HTTP: `GET/POST /movies`, `GET /movies/{id}`, `GET /health`
- PostgreSQL `movie_db`, таблица создаётся при старте
- Redis кэш для `GET /movies/{id}` (заголовок `X-Cache: HIT` / `MISS`)
- NATS событие `movie.created` при создании фильма

## Остановка

```bash
cd movie-service
docker compose down
```
