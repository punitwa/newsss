#!/bin/bash

echo "🚀 Starting Modern News Aggregator..."
echo ""

# Check if Go news server is running
if ! curl -s http://localhost:8082/health > /dev/null 2>&1; then
    echo "⚠️  Go news server is not running!"
    echo "💡 Starting the Go server..."
    
    # Kill any existing news server processes
    pkill -f "go run news-server.go" 2>/dev/null
    
    # Start Go server in background
    nohup go run news-server.go > /tmp/news-server.log 2>&1 &
    GO_PID=$!
    
    echo "⏳ Waiting for Go server to start..."
    sleep 8
    
    # Check if it started successfully
    if curl -s http://localhost:8082/health > /dev/null 2>&1; then
        echo "✅ Go news server started successfully!"
    else
        echo "❌ Failed to start Go server. Check /tmp/news-server.log for details"
        exit 1
    fi
else
    echo "✅ Go news server is already running!"
fi

echo ""
echo "🎨 Starting Modern React App..."

# Navigate to modern app directory
cd news-react-modern

# Start React app
echo "🌐 Starting React app on http://localhost:3001"
echo "📡 Connected to Go API on http://localhost:8082"
echo ""
echo "✨ Modern Features:"
echo "   • Tailwind CSS styling"
echo "   • Responsive design"
echo "   • Smooth animations"
echo "   • Modern UI components"
echo "   • Real-time search"
echo ""
echo "Press Ctrl+C to stop both servers"
echo ""

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "🛑 Stopping servers..."
    if [ ! -z "$GO_PID" ]; then
        kill $GO_PID 2>/dev/null
    fi
    pkill -f "go run news-server.go" 2>/dev/null
    exit 0
}

# Set trap for cleanup
trap cleanup SIGINT SIGTERM

# Start Python server
if command -v python3 &> /dev/null; then
    python3 serve.py
elif command -v python &> /dev/null; then
    python serve.py
else
    echo "❌ Python not found!"
    echo "💡 Opening React app directly in browser..."
    
    # Try to open in browser directly
    if command -v open &> /dev/null; then
        echo "🔄 Opening dev-server.html in browser..."
        open dev-server.html
    else
        echo "🌐 Please open: $(pwd)/dev-server.html"
    fi
    
    echo "Press Enter to stop the Go server..."
    read
    cleanup
fi
