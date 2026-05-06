#!/bin/sh
set -e

echo "Running database migrations..."
/migrator up

echo "Starting server..."
exec /server
