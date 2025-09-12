#!/bin/bash

echo "🧪 Testing Enhanced News Collection"
echo "=================================="

echo "📊 Current Database Stats:"
./monitor.sh | grep -A 10 "Database Stats"

echo ""
echo "🔄 Triggering fresh collection with enhanced processing..."

# Stop services
pkill -f "data-collector" || true
pkill -f "processor" || true
sleep 3

# Start services
echo "Starting processor..."
go run ./cmd/processor > processor-enhanced.log 2>&1 &
PROCESSOR_PID=$!

sleep 5

echo "Starting data collector..."
go run ./cmd/data-collector > collector-enhanced.log 2>&1 &
COLLECTOR_PID=$!

echo "⏳ Waiting for collection and processing..."
sleep 30

echo ""
echo "📊 Testing Enhanced Articles:"
curl -s "http://localhost:8082/api/v1/news?page=1&limit=3" | jq '.data[] | {
  title: .title,
  summary_words: (.summary | split(" ") | length),
  content_words: (.content | split(" ") | length),
  has_image: (.image_url != "" and .image_url != null),
  source: .source
}'

echo ""
echo "🖼️ Image URLs (first 3):"
curl -s "http://localhost:8082/api/v1/news?page=1&limit=3" | jq '.data[] | select(.image_url != "" and .image_url != null) | {title: .title, image_url: .image_url}'

echo ""
echo "📈 Summary: Enhanced collection with 80+ word descriptions and image extraction"
echo "🛑 To stop: kill $PROCESSOR_PID $COLLECTOR_PID"
