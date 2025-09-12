#!/bin/bash

echo "📊 News Aggregator System Monitor"
echo "=================================="
echo ""

echo "🏥 System Health:"
curl -s http://localhost:8082/health | jq . || echo "API Gateway not running"
echo ""

echo "🐰 RabbitMQ Queues:"
docker-compose -f docker-compose.infrastructure.yml exec -T rabbitmq rabbitmqctl list_queues name messages consumers 2>/dev/null || echo "RabbitMQ not running"
echo ""

echo "📰 Database Stats:"
docker-compose -f docker-compose.infrastructure.yml exec -T postgres psql -U postgres -d news_aggregator -c "
SELECT 
    'Total Articles' as metric, 
    COUNT(*)::text as value 
FROM news
UNION ALL
SELECT 
    'Categories', 
    COUNT(DISTINCT category)::text 
FROM news
UNION ALL
SELECT 
    'Sources', 
    COUNT(DISTINCT source)::text 
FROM news;
" 2>/dev/null || echo "Database not ready or no data yet"

echo ""
echo "📋 Recent Articles:"
docker-compose -f docker-compose.infrastructure.yml exec -T postgres psql -U postgres -d news_aggregator -c "
SELECT 
    LEFT(title, 50) || '...' as title,
    source,
    category,
    published_at::date
FROM news 
ORDER BY published_at DESC 
LIMIT 5;
" 2>/dev/null || echo "No articles found"

echo ""
echo "🌐 Access URLs:"
echo "   • RabbitMQ: http://localhost:15672 (guest/guest)"
echo "   • Grafana: http://localhost:3000 (admin/admin)"
echo "   • Prometheus: http://localhost:9090"
echo "   • Health Check: http://localhost:8082/health"
