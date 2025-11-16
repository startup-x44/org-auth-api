#!/bin/bash
echo "🔨 Rebuilding auth-service..."
docker-compose up -d --build auth-service

echo "⏳ Waiting for service to start..."
sleep 5

echo "📋 Recent logs:"
docker-compose logs auth-service --tail=15

echo "✅ Done! Service rebuilt and running"
