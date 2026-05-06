#!/bin/sh
set -e

echo "Running database migrations..."
if /migrator up; then
  echo "Migrations completed successfully"
else
  echo "Migration failed, but continuing to start server..."
fi

echo "Starting server on port ${PORT:-8080}..."
exec /server
