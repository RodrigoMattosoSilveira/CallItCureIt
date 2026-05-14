#!/usr/bin/env sh
set -eu

DB_PATH="${DATABASE_PATH:-/app/data/app.db}"

mkdir -p "$(dirname "$DB_PATH")"

echo "Applying database migrations to ${DB_PATH}..."

for migration in /app/migrations/*.up.sql; do
  echo "Applying ${migration}"
  sqlite3 "$DB_PATH" < "$migration"
done

echo "Starting API..."
exec /app/call-it-cure-it-api