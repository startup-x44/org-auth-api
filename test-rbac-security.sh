#!/bin/bash

# RBAC Security Test Runner
# This script runs the RBAC security tests with proper database configuration

set -e

echo "🔒 Running RBAC Security Tests..."
echo "=================================="

# Set test database configuration
export DB_HOST=${DB_HOST:-localhost}
export DB_PORT=${DB_PORT:-5432} 
export DB_USER=${DB_USER:-postgres}
export DB_PASSWORD=${DB_PASSWORD:-password}

echo "Database Configuration:"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  User: $DB_USER"
echo ""

# Check if PostgreSQL is available
if ! pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER >/dev/null 2>&1; then
    echo "❌ PostgreSQL is not available at $DB_HOST:$DB_PORT"
    echo "   Please ensure PostgreSQL is running and accessible"
    echo "   You can start it with: brew services start postgresql"
    exit 1
fi

echo "✅ PostgreSQL is available"
echo ""

# Run the RBAC security tests
echo "🧪 Running RBAC security tests..."
go test -v ./tests/feature -run TestRBAC

if [ $? -eq 0 ]; then
    echo ""
    echo "🎉 All RBAC security tests passed!"
    echo "🔒 Your RBAC system is secure and production-ready"
else
    echo ""
    echo "❌ Some RBAC security tests failed"
    echo "   Please review the test output above"
    exit 1
fi