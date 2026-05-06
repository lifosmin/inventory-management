#!/bin/sh
set -e

echo "Running database migrations..."
if /migrator up; then
  echo "Migrations completed successfully"
else
  echo "Migration failed, but continuing to start server..."
fi

echo "Seeding default admin user..."
if /seeder; then
  echo "Seed completed"
else
  echo "Seed skipped (user may already exist)"
fi

echo "Starting server on port ${PORT:-8080}..."
exec /server
