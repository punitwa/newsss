#!/bin/bash

echo "🧹 Testing Cleanup Functionality"
echo "================================="
echo ""

echo "📊 Current log file sizes:"
for file in api-gateway.log processor.log collector.log; do
    if [ -f "$file" ]; then
        size=$(wc -l < "$file" 2>/dev/null)
        size_mb=$(du -m "$file" 2>/dev/null | cut -f1)
        echo "   $file: $size lines (${size_mb}MB)"
    else
        echo "   $file: not found"
    fi
done
echo ""

echo "📈 Current database article count:"
PGPASSWORD=postgres psql -h localhost -p 5433 -U postgres -d news_aggregator -c "
SELECT 
    'Total articles' as category, COUNT(*) as count 
FROM news 
UNION ALL
SELECT 
    'Last 2 days' as category, COUNT(*) as count 
FROM news 
WHERE published_at >= NOW() - INTERVAL '2 days'
UNION ALL
SELECT 
    'Older than 2 days' as category, COUNT(*) as count 
FROM news 
WHERE published_at < NOW() - INTERVAL '2 days';
" 2>/dev/null
echo ""

echo "🧹 Testing database cleanup via API:"
response=$(curl -s -X POST http://localhost:8082/api/v1/admin/cleanup \
  -H "Authorization: Bearer your-jwt-token" 2>/dev/null)

if [ $? -eq 0 ]; then
    echo "   API Response: $response"
else
    echo "   ❌ API call failed (make sure API Gateway is running)"
fi
echo ""

echo "🔍 Testing log status check via API:"
response=$(curl -s -X POST http://localhost:8082/api/v1/admin/cleanup/logs \
  -H "Authorization: Bearer your-jwt-token" 2>/dev/null)

if [ $? -eq 0 ]; then
    echo "   API Response: $response"
else
    echo "   ❌ API call failed (make sure API Gateway is running)"
fi
echo ""

echo "📊 Database article count after cleanup:"
PGPASSWORD=postgres psql -h localhost -p 5433 -U postgres -d news_aggregator -c "
SELECT 
    'Total articles' as category, COUNT(*) as count 
FROM news 
UNION ALL
SELECT 
    'Last 2 days' as category, COUNT(*) as count 
FROM news 
WHERE published_at >= NOW() - INTERVAL '2 days';
" 2>/dev/null

echo ""
echo "✅ Cleanup test completed!"
echo ""
echo "💡 Notes:"
echo "   • Log rotation happens automatically when files exceed 10,000 lines"
echo "   • Database cleanup removes articles older than 2 days"
echo "   • Cleanup service runs every 6 hours automatically"
echo "   • Manual cleanup can be triggered via API endpoints"
