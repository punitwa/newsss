#!/bin/bash

echo "🚀 Starting React News App..."
echo ""

# Check if Go news server is running
if ! curl -s http://localhost:8082/health > /dev/null 2>&1; then
    echo "⚠️  Go news server is not running!"
    echo "💡 Please start it first with: go run news-server.go"
    echo ""
    read -p "Do you want to start the Go server now? (y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "🔄 Starting Go news server..."
        go run news-server.go &
        GO_PID=$!
        echo "⏳ Waiting for server to start..."
        sleep 5
    else
        echo "❌ React app needs the Go server to work. Exiting..."
        exit 1
    fi
fi

echo "✅ Go news server is running!"
echo ""

# Start React app server
cd news-react-app

echo "🌐 Starting React app on http://localhost:3001"
echo "📡 Connecting to Go API on http://localhost:8082"
echo ""
echo "Press Ctrl+C to stop"
echo ""

# Try Python 3 first, then Python
if command -v python3 &> /dev/null; then
    python3 serve.py
elif command -v python &> /dev/null; then
    python serve.py
else
    echo "❌ Python not found!"
    echo "💡 Please install Python or open news-react-app/index.html directly in your browser"
    echo "🌐 File location: $(pwd)/index.html"
    
    # Try to open in browser directly
    if command -v open &> /dev/null; then
        echo "🔄 Opening in browser..."
        open index.html
    fi
fi
