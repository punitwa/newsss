#!/bin/bash

echo "🚀 Opening News Aggregator Web Interface..."
echo ""
echo "📱 Your news website will open in your browser:"
echo "   http://localhost:8082"
echo ""
echo "🔧 If it doesn't open automatically, copy and paste the URL above"
echo ""

# Try to open in default browser
if command -v open &> /dev/null; then
    # macOS
    open http://localhost:8082
elif command -v xdg-open &> /dev/null; then
    # Linux
    xdg-open http://localhost:8082
elif command -v start &> /dev/null; then
    # Windows
    start http://localhost:8082
else
    echo "Please manually open: http://localhost:8082"
fi

echo "✅ Done!"
