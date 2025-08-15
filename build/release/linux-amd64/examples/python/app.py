#!/usr/bin/env python3
"""
{{.ProjectName}} - NSM Python Flask Example
{{.Description}}
"""

import json
import os
from datetime import datetime, timezone
from pathlib import Path
from typing import Dict, Any, Optional

from flask import Flask, request, jsonify, render_template_string
from flask_cors import CORS

# NSM Configuration
def load_nsm_config() -> Dict[str, Any]:
    """Load NSM port configuration if available."""
    default_config = {
        'http': {{.Port}},
        'https': {{.HTTPSPort}},
        'host': '127.0.0.1'
    }
    
    try:
        config_path = Path('.nsm-ports.json')
        if config_path.exists():
            with open(config_path, 'r') as f:
                config = json.load(f)
                print(f"🔧 NSM: Using HTTP port {config.get('http', default_config['http'])}")
                return {**default_config, **config}
    except Exception as e:
        print(f"NSM: Failed to load configuration: {e}")
    
    return default_config

# Initialize Flask app
app = Flask(__name__)
CORS(app)  # Enable CORS for development

# Load configuration
config = load_nsm_config()

# HTML template for the home page
HOME_TEMPLATE = """
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{project_name}} - NSM Python Example</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
            max-width: 900px;
            margin: 0 auto;
            padding: 2rem;
            line-height: 1.6;
            color: #374151;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
        }
        .container {
            background: white;
            border-radius: 16px;
            padding: 3rem;
            box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
        }
        .header {
            text-align: center;
            margin-bottom: 2rem;
        }
        .title {
            font-size: 2.5rem;
            font-weight: 700;
            color: #1f2937;
            margin-bottom: 0.5rem;
        }
        .subtitle {
            color: #6b7280;
            font-size: 1.1rem;
        }
        .python-badge {
            display: inline-block;
            background: linear-gradient(135deg, #3776ab 0%, #ffd43b 100%);
            color: white;
            padding: 0.25rem 0.75rem;
            border-radius: 12px;
            font-size: 0.8rem;
            font-weight: 600;
            margin: 0.5rem;
        }
        .feature-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
            gap: 1.5rem;
            margin: 2rem 0;
        }
        .feature {
            background: #fef7ff;
            padding: 1.5rem;
            border-radius: 12px;
            border: 2px solid #e0e7ff;
            transition: all 0.2s ease;
        }
        .feature:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(102, 126, 234, 0.2);
        }
        .feature h3 {
            margin: 0 0 0.5rem 0;
            color: #3730a3;
            font-size: 1.1rem;
        }
        .feature p {
            margin: 0;
            color: #6b7280;
            font-size: 0.9rem;
        }
        .api-section {
            background: #1f2937;
            color: white;
            padding: 2rem;
            border-radius: 12px;
            margin: 2rem 0;
        }
        .endpoint {
            background: rgba(255, 255, 255, 0.1);
            padding: 1rem;
            border-radius: 8px;
            margin: 1rem 0;
            font-family: 'SF Mono', Monaco, monospace;
            font-size: 0.9rem;
        }
        .method {
            display: inline-block;
            padding: 0.2rem 0.5rem;
            border-radius: 4px;
            font-size: 0.8rem;
            font-weight: bold;
            margin-right: 0.5rem;
        }
        .get { background: #10b981; }
        .post { background: #3b82f6; }
        .interactive-section {
            background: #f8fafc;
            padding: 2rem;
            border-radius: 12px;
            margin: 2rem 0;
            border: 1px solid #e2e8f0;
        }
        .form-demo {
            margin-top: 1rem;
        }
        .form-group {
            margin-bottom: 1rem;
        }
        .form-group label {
            display: block;
            margin-bottom: 0.5rem;
            font-weight: 600;
            color: #374151;
        }
        .form-group input, .form-group textarea {
            width: 100%;
            padding: 0.75rem;
            border: 2px solid #e2e8f0;
            border-radius: 8px;
            font-size: 1rem;
            box-sizing: border-box;
        }
        .form-group button {
            padding: 0.75rem 1.5rem;
            background: linear-gradient(135deg, #3776ab 0%, #ffd43b 100%);
            color: white;
            border: none;
            border-radius: 8px;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.2s ease;
        }
        .form-group button:hover {
            transform: translateY(-1px);
            box-shadow: 0 4px 12px rgba(55, 118, 171, 0.4);
        }
        .response {
            margin-top: 1rem;
            padding: 1rem;
            background: #f0f9ff;
            border-radius: 8px;
            border-left: 4px solid #0ea5e9;
            font-family: 'SF Mono', Monaco, monospace;
            font-size: 0.9rem;
            white-space: pre-wrap;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 class="title">🐍 {{project_name}}</h1>
            <p class="subtitle">Python Flask Web Server with NSM</p>
            <span class="python-badge">🐍 Python + Flask</span>
            <span class="python-badge">🚀 NSM Enabled</span>
        </div>

        <div class="feature-grid">
            <div class="feature">
                <h3>🐍 Python Power</h3>
                <p>Rapid development with Python's simplicity</p>
            </div>
            <div class="feature">
                <h3>🌐 Flask Framework</h3>
                <p>Lightweight and flexible web framework</p>
            </div>
            <div class="feature">
                <h3>🔧 Development Ready</h3>
                <p>Hot reload with Flask's debug mode</p>
            </div>
            <div class="feature">
                <h3>🔄 Auto Reload</h3>
                <p>Changes detected automatically</p>
            </div>
            <div class="feature">
                <h3>🌍 Custom Domain</h3>
                <p>Running on {{domain}} with HTTPS</p>
            </div>
            <div class="feature">
                <h3>📦 Package Management</h3>
                <p>Requirements.txt for dependencies</p>
            </div>
        </div>

        <div class="interactive-section">
            <h3>🧪 Interactive API Demo</h3>
            <p>Test the message endpoint:</p>
            <div class="form-demo">
                <div class="form-group">
                    <label for="messageInput">Message:</label>
                    <input type="text" id="messageInput" placeholder="Enter your message..." value="Hello from Python Flask!">
                </div>
                <div class="form-group">
                    <label for="authorInput">Author:</label>
                    <input type="text" id="authorInput" placeholder="Your name..." value="Flask Developer">
                </div>
                <div class="form-group">
                    <button onclick="testMessage()">Send Message</button>
                </div>
            </div>
            <div id="messageResponse" class="response" style="display: none;"></div>
        </div>

        <div class="api-section">
            <h3>🔗 API Endpoints</h3>
            <div class="endpoint">
                <span class="method get">GET</span>/api/info - Application information
            </div>
            <div class="endpoint">
                <span class="method get">GET</span>/api/health - Health check
            </div>
            <div class="endpoint">
                <span class="method post">POST</span>/api/message - Process message (JSON)
            </div>
            <div class="endpoint">
                <span class="method get">GET</span>/ - This page
            </div>
            <div class="endpoint">
                <span class="method get">GET</span>/static/&lt;path&gt; - Static file serving
            </div>
        </div>
    </div>

    <script>
        async function testMessage() {
            const messageInput = document.getElementById('messageInput');
            const authorInput = document.getElementById('authorInput');
            const responseDiv = document.getElementById('messageResponse');
            
            try {
                const response = await fetch('/api/message', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({ 
                        message: messageInput.value,
                        author: authorInput.value
                    })
                });
                
                const data = await response.json();
                responseDiv.textContent = JSON.stringify(data, null, 2);
                responseDiv.style.display = 'block';
            } catch (error) {
                responseDiv.textContent = 'Error: ' + error.message;
                responseDiv.style.display = 'block';
            }
        }

        // Allow Enter key to send message
        document.getElementById('messageInput').addEventListener('keypress', function(e) {
            if (e.key === 'Enter') {
                testMessage();
            }
        });
    </script>
</body>
</html>
"""

@app.route('/')
def home():
    """Home page with interactive demo."""
    return render_template_string(HOME_TEMPLATE, 
                                project_name="{{.ProjectName}}", 
                                domain="{{.Domain}}")

@app.route('/api/info')
def api_info():
    """Application information endpoint."""
    # Collect headers for debugging
    headers = {}
    if request.args.get('debug'):
        headers = dict(request.headers)
    
    nsm_enabled = os.getenv('NSM_ENABLED', 'false').lower() == 'true'
    
    info = {
        'name': '{{.ProjectName}}',
        'version': '1.0.0',
        'domain': '{{.Domain}}',
        'nsm_enabled': nsm_enabled,
        'python_version': f"{os.sys.version_info.major}.{os.sys.version_info.minor}.{os.sys.version_info.micro}",
        'flask_version': '3.0.0',
        'timestamp': datetime.now(timezone.utc).isoformat(),
    }
    
    if headers:
        info['headers'] = headers
    
    return jsonify(info)

@app.route('/api/health')
def health():
    """Health check endpoint."""
    return jsonify({
        'status': 'healthy',
        'timestamp': datetime.now(timezone.utc).isoformat(),
        'uptime': 'running'
    })

@app.route('/api/message', methods=['POST'])
def process_message():
    """Process a message and return response."""
    try:
        data = request.get_json()
        
        if not data or 'message' not in data:
            return jsonify({'error': 'Message is required'}), 400
        
        response = {
            'processed_message': data['message'].upper(),
            'original_message': data['message'],
            'author': data.get('author', 'Anonymous'),
            'length': len(data['message']),
            'word_count': len(data['message'].split()),
            'timestamp': datetime.now(timezone.utc).isoformat(),
            'id': f"msg_{datetime.now().timestamp():.0f}"
        }
        
        return jsonify(response)
    
    except Exception as e:
        return jsonify({'error': str(e)}), 500

@app.errorhandler(404)
def not_found(error):
    """404 error handler."""
    return jsonify({
        'error': 'Not Found',
        'message': 'The requested resource was not found',
        'timestamp': datetime.now(timezone.utc).isoformat()
    }), 404

@app.errorhandler(500)
def internal_error(error):
    """500 error handler."""
    return jsonify({
        'error': 'Internal Server Error',
        'message': 'An unexpected error occurred',
        'timestamp': datetime.now(timezone.utc).isoformat()
    }), 500

if __name__ == '__main__':
    print(f"🚀 Python Flask server starting on {config['host']}:{config['http']}")
    print(f"🌐 Domain: {{.Domain}}")
    print(f"📡 NSM: {'Enabled' if os.getenv('NSM_ENABLED') == 'true' else 'Disabled'}")
    print(f"🐍 Python: {os.sys.version_info.major}.{os.sys.version_info.minor}.{os.sys.version_info.micro}")
    print()
    
    app.run(
        host=config['host'],
        port=config['http'],
        debug=True,
        use_reloader=True
    )
