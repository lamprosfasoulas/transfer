#!/usr/bin/env ash
set -e

echo "Database is not used anymore..."
echo "Skipping migrations..."

#echo $DB_URL
#/usr/local/bin/migrate -path /app/migrations -database $DB_URL up

echo "Starting application... "

exec ./main
