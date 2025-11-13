#!/bin/bash

# Test runner script for auth-service
# This script sets up the test environment and runs all tests

set -e

echo "🚀 Starting auth-service tests..."

# Set test environment variables
export TEST_DB_HOST=${TEST_DB_HOST:-localhost}
export TEST_DB_PORT=${TEST_DB_PORT:-5432}
export TEST_DB_USER=${TEST_DB_USER:-auth_user}
export TEST_DB_PASSWORD=${TEST_DB_PASSWORD:-auth_password}
export TEST_DB_SSLMODE=${TEST_DB_SSLMODE:-disable}

echo "📊 Test Database: $TEST_DB_HOST:$TEST_DB_PORT"

# Wait for database to be ready
echo "⏳ Waiting for database to be ready..."
for i in {1..30}; do
    if pg_isready -h "$TEST_DB_HOST" -p "$TEST_DB_PORT" -U "$TEST_DB_USER" 2>/dev/null; then
        echo "✅ Database is ready!"
        break
    fi
    echo "Waiting for database... ($i/30)"
    sleep 2
done

if ! pg_isready -h "$TEST_DB_HOST" -p "$TEST_DB_PORT" -U "$TEST_DB_USER" 2>/dev/null; then
    echo "❌ Database connection failed"
    exit 1
fi

# Run tests
echo "🧪 Running unit tests..."
go test ./tests/unit/... -v

echo "🎯 Running feature tests..."
go test ./tests/feature/... -v

echo "✅ All tests completed successfully!"