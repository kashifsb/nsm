package com.nsm.example.controller;

import org.springframework.http.MediaType;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.ResponseBody;

@Controller
public class HomeController {

    @GetMapping(value = "/", produces = MediaType.TEXT_HTML_VALUE)
    @ResponseBody
    public String home() {
        return """
            <!DOCTYPE html>
            <html lang="en">
            <head>
                <meta charset="UTF-8">
                <meta name="viewport" content="width=device-width, initial-scale=1.0">
                <title>{{.ProjectName}} - NSM Java Example</title>
                <style>
                    body {
                        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
                        max-width: 900px;
                        margin: 0 auto;
                        padding: 2rem;
                        line-height: 1.6;
                        color: #374151;
                        background: linear-gradient(135deg, #ff7b7b 0%, #667eea 100%);
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
                    .java-badge {
                        display: inline-block;
                        background: linear-gradient(135deg, #ff7b7b 0%, #667eea 100%);
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
                        border: 2px solid #fde2e2;
                        transition: all 0.2s ease;
                    }
                    .feature:hover {
                        transform: translateY(-2px);
                        box-shadow: 0 8px 25px rgba(255, 123, 123, 0.2);
                    }
                    .feature h3 {
                        margin: 0 0 0.5rem 0;
                        color: #dc2626;
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
                    .test-button {
                        padding: 0.75rem 1.5rem;
                        background: linear-gradient(135deg, #ff7b7b 0%, #667eea 100%);
                        color: white;
                        border: none;
                        border-radius: 8px;
                        font-weight: 600;
                        cursor: pointer;
                        transition: all 0.2s ease;
                        margin: 0.5rem;
                    }
                    .test-button:hover {
                        transform: translateY(-1px);
                        box-shadow: 0 4px 12px rgba(255, 123, 123, 0.4);
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
                        <h1 class="title">☕ {{.ProjectName}}</h1>
                        <p class="subtitle">Java Spring Boot Web Server with NSM</p>
                        <span class="java-badge">☕ Java + Spring Boot</span>
                        <span class="java-badge">🚀 NSM Enabled</span>
                    </div>

                    <div class="feature-grid">
                        <div class="feature">
                            <h3>☕ Enterprise Java</h3>
                            <p>Robust, scalable, and production-ready</p>
                        </div>
                        <div class="feature">
                            <h3>🍃 Spring Boot</h3>
                            <p>Modern Java framework with auto-configuration</p>
                        </div>
                        <div class="feature">
                            <h3>🔧 Development Tools</h3>
                            <p>Hot reload with Spring DevTools</p>
                        </div>
                        <div class="feature">
                            <h3>📊 Monitoring</h3>
                            <p>Built-in actuator endpoints</p>
                        </div>
                        <div class="feature">
                            <h3>🌍 Custom Domain</h3>
                            <p>Running on {{.Domain}} with HTTPS</p>
                        </div>
                        <div class="feature">
                            <h3>🏗️ Maven Build</h3>
                            <p>Dependency management and build automation</p>
                        </div>
                    </div>

                    <div class="interactive-section">
                        <h3>🧪 Interactive API Demo</h3>
                        <p>Test the API endpoints:</p>
                        <button class="test-button" onclick="testEndpoint('/api/info')">Test Info API</button>
                        <button class="test-button" onclick="testEndpoint('/api/task', 'POST')">Test Task API</button>
                        <button class="test-button" onclick="testEndpoint('/actuator/health')">Test Health</button>
                        <div id="apiResponse" class="response" style="display: none;"></div>
                    </div>

                    <div class="api-section">
                        <h3>🔗 API Endpoints</h3>
                        <div class="endpoint">
                            <span class="method get">GET</span>/api/info - Application information
                        </div>
                        <div class="endpoint">
                            <span class="method post">POST</span>/api/task - Create task (JSON)
                        </div>
                        <div class="endpoint">
                            <span class="method get">GET</span>/actuator/health - Spring Boot health check
                        </div>
                        <div class="endpoint">
                            <span class="method get">GET</span>/actuator/info - Application info
                        </div>
                        <div class="endpoint">
                            <span class="method get">GET</span>/ - This page
                        </div>
                    </div>
                </div>

                <script>
                    async function testEndpoint(endpoint, method = 'GET') {
                        const responseDiv = document.getElementById('apiResponse');
                        
                        try {
                            let requestOptions = {
                                method: method,
                                headers: {
                                    'Content-Type': 'application/json',
                                }
                            };
                            
                            if (method === 'POST' && endpoint === '/api/task') {
                                requestOptions.body = JSON.stringify({
                                    title: 'Test Task',
                                    description: 'Task created from web interface',
                                    priority: 'HIGH'
                                });
                            }
                            
                            const response = await fetch(endpoint, requestOptions);
                            const data = await response.json();
                            responseDiv.textContent = JSON.stringify(data, null, 2);
                            responseDiv.style.display = 'block';
                        } catch (error) {
                            responseDiv.textContent = 'Error: ' + error.message;
                            responseDiv.style.display = 'block';
                        }
                    }
                </script>
            </body>
            </html>
            """;
    }
}
