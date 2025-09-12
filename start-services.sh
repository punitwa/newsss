gres#!/bin/bash

echo "🚀 Starting Complete News Aggregator System"
echo "============================================"

# Stop any running processes
echo "🛑 Stopping existing services..."
pkill -f "go run ./cmd/data-collector" 2>/dev/null || true
pkill -f "go run ./cmd/processor" 2>/dev/null || true
pkill -f "go run ./cmd/api-gateway" 2>/dev/null || true

# Start infrastructure
echo "📦 Starting infrastructure..."
docker-compose -f docker-compose.infrastructure.yml up -d

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 30

# Check service health
echo "🔍 Checking service health..."
until docker-compose -f docker-compose.infrastructure.yml exec -T postgres pg_isready -U postgres >/dev/null 2>&1; do
    echo "   Waiting for PostgreSQL..."
    sleep 2
done
echo "✅ PostgreSQL is ready"

until docker-compose -f docker-compose.infrastructure.yml exec -T redis redis-cli ping >/dev/null 2>&1; do
    echo "   Waiting for Redis..."
    sleep 2
done
echo "✅ Redis is ready"

until curl -s http://localhost:15672 >/dev/null 2>&1; do
    echo "   Waiting for RabbitMQ..."
    sleep 2
done
echo "✅ RabbitMQ is ready"

# Create database table
echo "🗄️ Setting up database..."
docker-compose -f docker-compose.infrastructure.yml exec -T postgres psql -U postgres -d news_aggregator -c "
CREATE TABLE IF NOT EXISTS news (
    id VARCHAR PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT,
    summary TEXT,
    url TEXT UNIQUE,
    image_url TEXT,
    author TEXT,
    source TEXT,
    category TEXT,
    tags TEXT[],
    published_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    hash TEXT UNIQUE
);
"

# Start services
echo "🚀 Starting application services..."
echo "   Starting API Gateway..."
go run ./cmd/api-gateway &
API_PID=$!

sleep 5

echo "   Starting Data Collector..."
go run ./cmd/data-collector &
COLLECTOR_PID=$!

sleep 5

echo "   Starting Processor..."
go run ./cmd/processor &
PROCESSOR_PID=$!

sleep 3

echo "   Starting Cleanup Service..."
go run ./cmd/cleanup &
CLEANUP_PID=$!

echo ""
echo "✅ All services started!"
echo ""
echo "📊 Monitor your system:"
echo "   • ./monitor.sh"
echo "   • RabbitMQ: http://localhost:15672 (guest/guest)"
echo "   • API Health: http://localhost:8082/health"
echo "   • React App: http://localhost:3001"
echo ""
echo "🛑 To stop all services:"
echo "   kill $API_PID $COLLECTOR_PID $PROCESSOR_PID"
echo "   docker-compose -f docker-compose.infrastructure.yml down"
