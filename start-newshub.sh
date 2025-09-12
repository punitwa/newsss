#!/bin/bash

echo "🎨 Starting NewsHub - Professional News Interface"
echo ""

# Check if Go news server is running
if ! curl -s http://localhost:8082/health > /dev/null 2>&1; then
    echo "⚠️  Go news server is not running!"
    echo "🚀 Starting the Go server..."
    
    # Kill any existing news server processes
    pkill -f "go run news-server.go" 2>/dev/null
    
    # Start Go server in background
    nohup go run news-server.go > /tmp/news-server.log 2>&1 &
    
    echo "⏳ Waiting for Go server to start..."
    sleep 8
    
    # Check if it started successfully
    if curl -s http://localhost:8082/health > /dev/null 2>&1; then
        echo "✅ Go news server started!"
    else
        echo "❌ Failed to start Go server. Check /tmp/news-server.log"
        exit 1
    fi
else
    echo "✅ Go news server is running!"
fi

echo ""
echo "🎨 Starting NewsHub Interface..."
echo ""
echo "📱 NewsHub will open at: http://localhost:3003"
echo "📡 Connected to API: http://localhost:8082"
echo ""
echo "✨ NewsHub Features:"
echo "   • Professional blue gradient design"
echo "   • Hero section with stats"
echo "   • Top Stories grid"
echo "   • Latest News horizontal cards"
echo "   • Trending topics sidebar"
echo "   • Real-time search"
echo "   • Category navigation"
echo ""
echo "Press Ctrl+C to stop"
echo ""

# Start NewsHub
cd news-hub-app

if command -v python3 &> /dev/null; then
    python3 serve.py
elif command -v python &> /dev/null; then
    python serve.py
else
    echo "❌ Python not found!"
    echo "🌐 Opening NewsHub directly in browser..."
    
    if command -v open &> /dev/null; then
        open index.html
    else
        echo "💡 Please open: $(pwd)/index.html"
    fi
    
    echo "Press Enter to continue..."
    read
fi
