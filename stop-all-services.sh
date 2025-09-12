#!/bin/bash

# News Aggregator - Stop All Services Script
# This script stops all running services

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

echo "╔══════════════════════════════════════╗"
echo "║     News Aggregator Services         ║"
echo "║           Stop Script                ║"
echo "╚══════════════════════════════════════╝"
echo ""

print_status "Stopping all News Aggregator services..."

# Stop Go services
pkill -f "data-collector" 2>/dev/null && echo "✅ Data Collector stopped" || echo "• Data Collector was not running"
pkill -f "processor" 2>/dev/null && echo "✅ Processor stopped" || echo "• Processor was not running"
pkill -f "api-gateway" 2>/dev/null && echo "✅ API Gateway stopped" || echo "• API Gateway was not running"

# Stop React frontend
pkill -f "npm run dev" 2>/dev/null && echo "✅ React Frontend stopped" || echo "• React Frontend was not running"
pkill -f "vite" 2>/dev/null || true

sleep 2

print_success "🛑 All services stopped successfully!"
echo ""
echo "💡 To start services again, run: ./start-all-services.sh"
