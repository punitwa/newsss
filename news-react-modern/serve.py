#!/usr/bin/env python3
import http.server
import socketserver
import webbrowser
import os
import sys

PORT = 3002
DIRECTORY = os.path.dirname(os.path.abspath(__file__))

class Handler(http.server.SimpleHTTPRequestHandler):
    def __init__(self, *args, **kwargs):
        super().__init__(*args, directory=DIRECTORY, **kwargs)
    
    def end_headers(self):
        self.send_header('Access-Control-Allow-Origin', '*')
        self.send_header('Access-Control-Allow-Methods', 'GET, POST, OPTIONS')
        self.send_header('Access-Control-Allow-Headers', 'Content-Type')
        super().end_headers()
    
    def do_GET(self):
        # Serve dev-server.html for root path
        if self.path == '/' or self.path == '/index.html':
            self.path = '/dev-server.html'
        super().do_GET()

if __name__ == "__main__":
    try:
        with socketserver.TCPServer(("", PORT), Handler) as httpd:
            print(f"🚀 Starting Modern React News App...")
            print(f"📱 Available at: http://localhost:{PORT}")
            print(f"🔧 Connecting to Go API: http://localhost:8082")
            print(f"")
            print(f"✨ Features:")
            print(f"   • Modern React with Tailwind CSS")
            print(f"   • Responsive design")
            print(f"   • Real-time search")
            print(f"   • Category filtering")
            print(f"   • Beautiful animations")
            print(f"")
            print(f"Press Ctrl+C to stop")
            print(f"")
            
            # Try to open browser
            try:
                webbrowser.open(f'http://localhost:{PORT}')
                print(f"✅ Browser opened automatically")
            except:
                print(f"💡 Please manually open: http://localhost:{PORT}")
            
            httpd.serve_forever()
    except KeyboardInterrupt:
        print(f"\n🛑 Server stopped")
        sys.exit(0)
    except OSError as e:
        if e.errno == 48:  # Address already in use
            print(f"❌ Port {PORT} is already in use")
            print(f"💡 Try stopping other servers or changing the port in serve.py")
        else:
            print(f"❌ Error starting server: {e}")
        sys.exit(1)
