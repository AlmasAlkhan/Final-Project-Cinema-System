#!/bin/bash
set -e

curl -s http://localhost:8080/health
echo ""

curl -s http://localhost:8080/movies
echo ""

curl -s -X POST http://localhost:8080/movies \
  -H "Content-Type: application/json" \
  -d '{"title":"Dune Part Two","description":"Epic sci-fi","duration_min":166,"poster_url":"dune2.jpg","release_date":"2024-03-01"}'
echo ""

echo "First request (expect X-Cache: MISS):"
curl -s -D - http://localhost:8080/movies/1 -o /dev/null | grep -i x-cache || true
curl -s http://localhost:8080/movies/1
echo ""

echo "Second request (expect X-Cache: HIT):"
curl -s -D - http://localhost:8080/movies/1 -o /dev/null | grep -i x-cache || true
curl -s http://localhost:8080/movies/1
echo ""
