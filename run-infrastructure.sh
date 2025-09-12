#!/bin/bash

echo "🚀 Starting News Aggregator Infrastructure..."
echo ""

# Start infrastructure services
echo "📦 Starting PostgreSQL, Redis, RabbitMQ, Elasticsearch..."
docker-compose -f docker-compose.infrastructure.yml up -d

echo ""
echo "⏳ Waiting for services to start..."
sleep 30

echo ""
echo "🔍 Checking service health..."

# Check PostgreSQL
if docker-compose -f docker-compose.infrastructure.yml exec -T postgres pg_isready -U postgres >/dev/null 2>&1; then
    echo "✅ PostgreSQL is ready"
else
    echo "❌ PostgreSQL is not ready"
fi

# Check Redis
if docker-compose -f docker-compose.infrastructure.yml exec -T redis redis-cli ping >/dev/null 2>&1; then
    echo "✅ Redis is ready"
else
    echo "❌ Redis is not ready"
fi

# Check Elasticsearch
if curl -s http://localhost:9200/_cluster/health >/dev/null 2>&1; then
    echo "✅ Elasticsearch is ready"
else
    echo "❌ Elasticsearch is not ready"
fi

# Check RabbitMQ
if curl -s http://localhost:15672 >/dev/null 2>&1; then
    echo "✅ RabbitMQ Management UI is ready"
else
    echo "❌ RabbitMQ Management UI is not ready"
fi

echo ""
echo "🎉 Infrastructure is running!"
echo ""
echo "📋 Access your services:"
echo "   • RabbitMQ Management: http://localhost:15672 (guest/guest)"
echo "   • Elasticsearch: http://localhost:9200"
echo "   • Grafana: http://localhost:3000 (admin/admin)"
echo "   • Prometheus: http://localhost:9090"
echo ""
echo "🔧 Next steps:"
echo "   1. Run: go run ./cmd/api-gateway"
echo "   2. Run: go run ./cmd/data-collector" 
echo "   3. Run: go run ./cmd/processor"
echo ""
echo "💡 To stop infrastructure: docker-compose -f docker-compose.infrastructure.yml down"
