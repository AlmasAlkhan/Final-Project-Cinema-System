#!/bin/bash
set -e
cd "$(dirname "$0")"
chmod +x movie-service/scripts/setup.sh movie-service/scripts/test-grpc.sh
./movie-service/scripts/setup.sh
