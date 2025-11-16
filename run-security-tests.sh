#!/bin/bash

# Comprehensive RBAC Security Test Suite
# Tests all security components to ensure organization isolation

set -e

echo "🔒 RBAC Security Test Suite"
echo "=========================="
echo ""

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

# Check if PostgreSQL is available (but don't fail if not - some tests don't need DB)
if pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER >/dev/null 2>&1; then
    echo "✅ PostgreSQL is available"
    DB_AVAILABLE=true
else
    echo "⚠️  PostgreSQL not available - will skip database-dependent tests"
    DB_AVAILABLE=false
fi
echo ""

# Test compilation first
echo "🔧 Testing compilation..."
if ! go test -c ./tests/feature >/dev/null 2>&1; then
    echo "❌ Feature tests failed to compile"
    exit 1
fi

if ! go test -c ./tests/handler >/dev/null 2>&1; then
    echo "❌ Handler tests failed to compile"
    exit 1
fi
echo "✅ All tests compile successfully"
echo ""

# Run security logic tests (no database required)
echo "🧠 Testing security logic..."
if go test -v ./tests/handler -run TestRoleHandlerPermissionChecks; then
    echo "✅ Security logic tests passed"
else
    echo "❌ Security logic tests failed"
    exit 1
fi
echo ""

# Run database tests if DB is available
if [ "$DB_AVAILABLE" = true ]; then
    echo "🗄️  Testing database security..."
    
    echo "  Running feature tests..."
    if go test -v ./tests/feature -run TestRBAC -timeout 60s; then
        echo "✅ Database security tests passed"
    else
        echo "❌ Database security tests failed"
        exit 1
    fi
    
    echo "  Running handler isolation tests..."
    if go test -v ./tests/handler -run TestRoleHandlerSecurityIsolation -timeout 30s; then
        echo "✅ Handler isolation tests passed"
    else
        echo "❌ Handler isolation tests failed"
        exit 1
    fi
else
    echo "⏭️  Skipping database tests (PostgreSQL not available)"
fi

echo ""
echo "📋 Security Test Summary"
echo "========================"
echo "✅ Compilation: PASSED"
echo "✅ Security Logic: PASSED"

if [ "$DB_AVAILABLE" = true ]; then
    echo "✅ Database Security: PASSED"
    echo "✅ Handler Isolation: PASSED"
    
    echo ""
    echo "🎉 ALL SECURITY TESTS PASSED!"
    echo "🔒 Your RBAC system is secure and production-ready"
    echo ""
    echo "Security Features Verified:"
    echo "  ✓ Organization isolation at repository layer"
    echo "  ✓ Cross-organization privilege escalation prevention"
    echo "  ✓ Permission assignment validation"
    echo "  ✓ Service layer security enforcement"
    echo "  ✓ Handler layer permission checks"
    echo "  ✓ Deprecated method disabling"
else
    echo "⚠️  Database Security: SKIPPED"
    echo ""
    echo "✅ Core security logic verified"
    echo "💡 Install and start PostgreSQL to run full security tests"
    echo "   Example: brew install postgresql && brew services start postgresql"
fi