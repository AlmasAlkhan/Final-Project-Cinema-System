#!/bin/bash
set -e

GRPCURL="$(go env GOPATH)/bin/grpcurl"
if [ ! -x "$GRPCURL" ]; then
  go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
  GRPCURL="$(go env GOPATH)/bin/grpcurl"
fi

"$GRPCURL" -plaintext localhost:50051 list

"$GRPCURL" -plaintext -d '{"title":"Inception","description":"Dream","duration_min":148,"poster_url":"inception.jpg","release_date":"2010-07-16"}' \
  localhost:50051 movie.MovieService/CreateMovie

"$GRPCURL" -plaintext -d '{"id":1}' localhost:50051 movie.MovieService/GetMovie
"$GRPCURL" -plaintext -d '{"id":1}' localhost:50051 movie.MovieService/GetMovie

"$GRPCURL" -plaintext -d '{"page":1,"page_size":10}' localhost:50051 movie.MovieService/ListMovies
"$GRPCURL" -plaintext -d '{}' localhost:50051 movie.MovieService/ListHalls

"$GRPCURL" -plaintext -d '{"movie_id":1,"hall_id":1,"start_time":"2026-05-20T18:00:00Z","duration_min":148,"price":2500}' \
  localhost:50051 movie.MovieService/CreateShowtime

"$GRPCURL" -plaintext -d '{"movie_id":1}' localhost:50051 movie.MovieService/ListShowtimes
"$GRPCURL" -plaintext -d '{"hall_id":1}' localhost:50051 movie.MovieService/GetHallLayout
