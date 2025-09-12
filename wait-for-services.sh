
#!/bin/bash

echo "⏳ Waiting for all services to be ready..."

# Wait for PostgreSQL
echo "🐘 Waiting for PostgreSQL..."
until docker-compose -f docker-compose.infrastructure.yml exec -T postgres pg_isready -U postgres >/dev/null 2>&1; do
    echo "   PostgreSQL not ready, waiting..."
    sleep 2
done
echo "✅ PostgreSQL is ready"

# Wait for RabbitMQ
echo "🐰 Waiting for RabbitMQ..."
until curl -s http://localhost:15672 >/dev/null 2>&1; do
    echo "   RabbitMQ not ready, waiting..."
    sleep 2
done
echo "✅ RabbitMQ is ready"

# Test database connection
echo "🔗 Testing database connection..."
until docker-compose -f docker-compose.infrastructure.yml exec -T postgres psql -U postgres -d news_aggregator -c "SELECT 1;" >/dev/null 2>&1; do
    echo "   Database not ready, waiting..."
    sleep 2
done
echo "✅ Database connection successful"

# Test RabbitMQ connection
echo "📨 Testing RabbitMQ connection..."
until curl -u guest:guest -s http://localhost:15672/api/overview >/dev/null 2>&1; do
    echo "   RabbitMQ API not ready, waiting..."
    sleep 2
done
echo "✅ RabbitMQ API ready"

echo ""
echo "🎉 All services are ready! You can now start:"
echo "   • go run ./cmd/api-gateway"
echo "   • go run ./cmd/data-collector" 
echo "   • go run ./cmd/processor"
