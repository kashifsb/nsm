package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type NSMConfig struct {
	HTTP  int    `json:"http"`
	HTTPS int    `json:"https"`
	Host  string `json:"host"`
}

type AppInfo struct {
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Domain    string            `json:"domain"`
	NSM       bool              `json:"nsm_enabled"`
	Timestamp time.Time         `json:"timestamp"`
	Headers   map[string]string `json:"headers,omitempty"`
}

func loadNSMConfig() NSMConfig {
	config := NSMConfig{
		HTTP:  8080,
		HTTPS: 8443,
		Host:  "127.0.0.1",
	}

	// Read NSM port configuration
	if data, err := os.ReadFile(".nsm-ports.json"); err == nil {
		if err := json.Unmarshal(data, &config); err != nil {
			log.Printf("NSM: Failed to parse configuration: %v", err)
		} else {
			log.Printf("🔧 NSM: Using HTTP port %d", config.HTTP)
		}
	}

	return config
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.ProjectName}} - NSM Go Example</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
            max-width: 800px;
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
        .feature-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 1.5rem;
            margin: 2rem 0;
        }
        .feature {
            background: #f8fafc;
            padding: 1.5rem;
            border-radius: 12px;
            border: 1px solid #e2e8f0;
        }
        .feature h3 {
            margin: 0 0 0.5rem 0;
            color: #374151;
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
        }
        .nsm-badge {
            display: inline-block;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 0.25rem 0.75rem;
            border-radius: 12px;
            font-size: 0.8rem;
            font-weight: 600;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 class="title">🚀 {{.ProjectName}}</h1>
            <p class="subtitle">Go Web Server with NSM</p>
            <span class="nsm-badge">NSM Enabled</span>
        </div>

        <div class="feature-grid">
            <div class="feature">
                <h3>🔧 Development Ready</h3>
                <p>Hot reload with automatic port configuration</p>
            </div>
            <div class="feature">
                <h3>🌐 Custom Domain</h3>
                <p>Running on {{.Domain}} with HTTPS</p>
            </div>
            <div class="feature">
                <h3>⚡ Fast Build</h3>
                <p>Go's lightning-fast compilation</p>
            </div>
            <div class="feature">
                <h3>🔒 Secure</h3>
                <p>Local HTTPS certificates</p>
            </div>
        </div>

        <div class="api-section">
            <h3>🔗 API Endpoints</h3>
            <div class="endpoint">GET /api/info - Application information</div>
            <div class="endpoint">GET /api/health - Health check</div>
            <div class="endpoint">GET / - This page</div>
        </div>
    </div>
</body>
</html>`

	fmt.Fprint(w, html)
}

func apiInfoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Collect headers for debugging
	headers := make(map[string]string)
	for name, values := range r.Header {
		if len(values) > 0 {
			headers[name] = values[0]
		}
	}

	info := AppInfo{
		Name:      "{{.ProjectName}}",
		Version:   "1.0.0",
		Domain:    "{{.Domain}}",
		NSM:       os.Getenv("NSM_ENABLED") == "true",
		Timestamp: time.Now(),
		Headers:   headers,
	}

	json.NewEncoder(w).Encode(info)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"uptime":    "running",
	}

	json.NewEncoder(w).Encode(response)
}

func main() {
	config := loadNSMConfig()

	// Setup routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/api/info", apiInfoHandler)
	http.HandleFunc("/api/health", healthHandler)

	// Serve static files
	fs := http.FileServer(http.Dir("./static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	addr := config.Host + ":" + strconv.Itoa(config.HTTP)

	fmt.Printf("🚀 Go server starting on %s\n", addr)
	fmt.Printf("🌐 Domain: {{.Domain}}\n")
	fmt.Printf("📡 NSM: %s\n", map[bool]string{true: "Enabled", false: "Disabled"}[os.Getenv("NSM_ENABLED") == "true"])
	fmt.Println()

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
