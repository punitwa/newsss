#!/bin/bash

echo "🚀 Starting Simple News Aggregator (Infrastructure Only)..."
echo ""

# Check if infrastructure is running
echo "🔍 Checking if infrastructure is running..."

if ! docker ps | grep -q news_postgres; then
    echo "❌ Infrastructure not running. Starting it first..."
    ./run-infrastructure.sh
    echo "⏳ Waiting for infrastructure to be ready..."
    sleep 30
fi

echo "✅ Infrastructure is running!"
echo ""

echo "📋 Available Services:"
echo "   • PostgreSQL: localhost:5432"
echo "   • Redis: localhost:6379"
echo "   • RabbitMQ: localhost:5672 (Management: http://localhost:15672)"
echo "   • Elasticsearch: http://localhost:9200"
echo "   • Grafana: http://localhost:3000 (admin/admin)"
echo "   • Prometheus: http://localhost:9090"
echo ""

echo "🎯 To test the services:"
echo ""
echo "Test PostgreSQL:"
echo "  docker-compose -f docker-compose.infrastructure.yml exec postgres psql -U postgres -d news_aggregator"
echo ""
echo "Test Redis:"
echo "  docker-compose -f docker-compose.infrastructure.yml exec redis redis-cli ping"
echo ""
echo "Test Elasticsearch:"
echo "  curl http://localhost:9200"
echo ""
echo "Test RabbitMQ Management:"
echo "  open http://localhost:15672 (guest/guest)"
echo ""

echo "💡 Next Steps:"
echo "1. Your infrastructure is ready!"
echo "2. You can now build a simple Go application that connects to these services"
echo "3. Or wait while we fix the dependency issues in the main application"
echo ""

echo "🛑 To stop infrastructure:"
echo "  docker-compose -f docker-compose.infrastructure.yml down"
