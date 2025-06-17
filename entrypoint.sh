#!/usr/bin/env sh
set -e

echo "Running database migrations... "

echo $DB_URL
/usr/local/bin/migrate -path /app/migrations -database $DB_URL up

echo "Starting application... "

exec ./main
