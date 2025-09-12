#!/bin/bash

# News Aggregator - Start All Services Script
# This script starts all required services in the correct order

set -e  # Exit on any error

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

echo "╔══════════════════════════════════════╗"
echo "║     News Aggregator Services         ║"
echo "║           Startup Script             ║"
echo "╚══════════════════════════════════════╝"
echo ""

# Check if we're in the right directory
if [ ! -f "configs/config.yaml" ]; then
    echo "❌ Error: Please run this script from the NEWSAggregator directory"
    echo "Current directory: $(pwd)"
    exit 1
fi

# Stop any existing services first
print_status "Stopping any existing services..."
pkill -f "data-collector" 2>/dev/null || true
pkill -f "processor" 2>/dev/null || true
pkill -f "api-gateway" 2>/dev/null || true
pkill -f "npm run dev" 2>/dev/null || true
pkill -f "vite" 2>/dev/null || true
sleep 2

# Check if Docker infrastructure is running
print_status "Checking Docker infrastructure..."
if ! docker ps | grep -q "news_postgres"; then
    print_status "Starting Docker infrastructure services..."
    docker-compose -f docker-compose.infrastructure.yml up -d
    print_status "Waiting for infrastructure to be ready..."
    sleep 15
else
    print_success "Docker infrastructure is already running"
fi

# Start Go services
print_status "Starting Processor..."
go run ./cmd/processor > processor.log 2>&1 &
PROCESSOR_PID=$!
sleep 3

print_status "Starting Data Collector..."
go run ./cmd/data-collector > collector.log 2>&1 &
COLLECTOR_PID=$!
sleep 3

print_status "Starting API Gateway..."
go run ./cmd/api-gateway > api-gateway.log 2>&1 &
GATEWAY_PID=$!
sleep 3

# Start React frontend
print_status "Starting React Frontend..."
cd news-react-modern
npm run dev > ../react-app.log 2>&1 &
REACT_PID=$!
cd ..
sleep 5

# Check if services are running
print_status "Checking service status..."
echo ""

if kill -0 $PROCESSOR_PID 2>/dev/null; then
    print_success "✅ Processor is running (PID: $PROCESSOR_PID)"
else
    echo "❌ Processor failed to start"
fi

if kill -0 $COLLECTOR_PID 2>/dev/null; then
    print_success "✅ Data Collector is running (PID: $COLLECTOR_PID)"
else
    echo "❌ Data Collector failed to start"
fi

if kill -0 $GATEWAY_PID 2>/dev/null; then
    print_success "✅ API Gateway is running (PID: $GATEWAY_PID)"
else
    echo "❌ API Gateway failed to start"
fi

if kill -0 $REACT_PID 2>/dev/null; then
    print_success "✅ React Frontend is running (PID: $REACT_PID)"
else
    echo "❌ React Frontend failed to start"
fi

echo ""
print_success "🎉 All services started successfully!"
echo ""
echo "📋 Service URLs:"
echo "• Frontend: http://localhost:5173/"
echo "• API: http://localhost:8082/"
echo "• Health Check: http://localhost:8082/health"
echo "• Grafana: http://localhost:3000/"
echo "• Prometheus: http://localhost:9090/"
echo ""
echo "📝 Process IDs:"
echo "• Processor: $PROCESSOR_PID"
echo "• Data Collector: $COLLECTOR_PID" 
echo "• API Gateway: $GATEWAY_PID"
echo "• React Frontend: $REACT_PID"
echo ""
print_warning "To stop all services, run: pkill -f 'processor|data-collector|api-gateway|vite'"
echo ""
print_status "Logs are available in:"
echo "• processor.log"
echo "• collector.log"
echo "• api-gateway.log"
echo "• react-app.log"
