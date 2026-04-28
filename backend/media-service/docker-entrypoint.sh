#!/bin/sh
set -e
echo "Running database migrations..."
/app/migrate
echo "Starting server..."
exec /app/server
