#!/bin/bash

# Development startup script for auth-service
# This script provides easy commands for development workflow

set -e

PROJECT_NAME="auth-service"
DEV_COMPOSE_FILE="docker-compose.dev.yml"
TEST_COMPOSE_FILE="docker-compose.test.yml"

case "$1" in
    "dev")
        echo "🚀 Starting development environment with live reload..."
        docker-compose -f $DEV_COMPOSE_FILE up --build
        ;;
    "dev-d")
        echo "🚀 Starting development environment in background..."
        docker-compose -f $DEV_COMPOSE_FILE up -d --build
        ;;
    "stop")
        echo "🛑 Stopping development environment..."
        docker-compose -f $DEV_COMPOSE_FILE down
        ;;
    "test")
        echo "🧪 Running tests..."
        docker-compose -f $TEST_COMPOSE_FILE up --abort-on-container-exit
        ;;
    "clean")
        echo "🧹 Cleaning up containers and volumes..."
        docker-compose -f $DEV_COMPOSE_FILE down -v
        docker-compose -f $TEST_COMPOSE_FILE down -v
        ;;
    "logs")
        echo "📋 Showing development logs..."
        docker-compose -f $DEV_COMPOSE_FILE logs -f
        ;;
    "shell")
        echo "🐚 Opening shell in auth-service container..."
        docker-compose -f $DEV_COMPOSE_FILE exec auth-service sh
        ;;
    *)
        echo "Usage: $0 {dev|dev-d|stop|test|clean|logs|shell}"
        echo ""
        echo "Commands:"
        echo "  dev     - Start development environment with live reload (foreground)"
        echo "  dev-d   - Start development environment in background"
        echo "  stop    - Stop development environment"
        echo "  test    - Run test suite"
        echo "  clean   - Clean up containers and volumes"
        echo "  logs    - Show development logs"
        echo "  shell   - Open shell in auth-service container"
        echo ""
        echo "Example: ./dev.sh dev"
        exit 1
        ;;
esac