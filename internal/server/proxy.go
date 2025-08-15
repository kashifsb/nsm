package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/kashifsb/nsm/internal/config"
	"github.com/kashifsb/nsm/pkg/logger"
)

type ProxyServer struct {
	cfg        *config.Config
	httpServer *http.Server
	targetURL  *url.URL
	certPath   string
	keyPath    string
}

type ProxyConfig struct {
	TargetHost  string
	TargetPort  int
	ProxyPort   int
	Domain      string
	CertPath    string
	KeyPath     string
	EnableHTTPS bool
}

func NewProxyServer(cfg *config.Config, proxyConfig ProxyConfig) *ProxyServer {
	targetURL, _ := url.Parse(fmt.Sprintf("http://%s:%d", proxyConfig.TargetHost, proxyConfig.TargetPort))

	return &ProxyServer{
		cfg:       cfg,
		targetURL: targetURL,
		certPath:  proxyConfig.CertPath,
		keyPath:   proxyConfig.KeyPath,
	}
}

func (p *ProxyServer) Start(ctx context.Context, port int) error {
	proxy := httputil.NewSingleHostReverseProxy(p.targetURL)

	// Enhanced proxy director
	originalDirector := proxy.Director
	proxy.Director = func(r *http.Request) {
		originalDirector(r)
		p.enhanceRequest(r)
	}

	// Custom error handler
	proxy.ErrorHandler = p.errorHandler

	// Create server
	p.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      p.middlewareChain(proxy),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if p.cfg.EnableHTTPS {
		p.httpServer.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS13,
		}
	}

	logger.Info("Starting proxy server",
		"port", port,
		"target", p.targetURL.String(),
		"https", p.cfg.EnableHTTPS)

	// Start server
	go func() {
		var err error
		if p.cfg.EnableHTTPS {
			err = p.httpServer.ListenAndServeTLS(p.certPath, p.keyPath)
		} else {
			err = p.httpServer.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			logger.Error("Proxy server error", "error", err)
		}
	}()

	// Wait for server to be ready
	return p.waitForReady(ctx, port)
}

func (p *ProxyServer) Stop(ctx context.Context) error {
	if p.httpServer == nil {
		return nil
	}

	logger.Info("Shutting down proxy server")
	return p.httpServer.Shutdown(ctx)
}

func (p *ProxyServer) enhanceRequest(r *http.Request) {
	// Set forwarded headers
	r.Header.Set("X-Forwarded-Proto", p.getScheme())
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Header.Set("X-Forwarded-For", p.getClientIP(r))

	// Set NSM headers
	r.Header.Set("X-NSM-Version", "3.0.0")
	r.Header.Set("X-NSM-Project", p.cfg.ProjectName)

	// Override host for backend
	r.Host = p.targetURL.Host
}

func (p *ProxyServer) middlewareChain(next http.Handler) http.Handler {
	// CORS middleware
	corsHandler := p.corsMiddleware(next)

	// Logging middleware
	loggingHandler := p.loggingMiddleware(corsHandler)

	// Recovery middleware
	recoveryHandler := p.recoveryMiddleware(loggingHandler)

	return recoveryHandler
}

func (p *ProxyServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for development
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (p *ProxyServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap ResponseWriter to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		logger.Debug("Request processed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration", duration,
			"remote_addr", p.getClientIP(r))
	})
}

func (p *ProxyServer) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic in proxy handler", "error", err, "path", r.URL.Path)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (p *ProxyServer) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	if strings.Contains(err.Error(), "connection refused") {
		p.renderDevServerNotReady(w, r)
		return
	}

	logger.Error("Proxy error", "error", err, "path", r.URL.Path)
	http.Error(w, "Bad Gateway", http.StatusBadGateway)
}

func (p *ProxyServer) renderDevServerNotReady(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusBadGateway)

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>NSM - Development Server Starting</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', system-ui, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 2rem;
            line-height: 1.6;
            color: #374151;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container {
            background: white;
            border-radius: 16px;
            padding: 3rem;
            box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
            text-align: center;
            max-width: 600px;
        }
        .logo {
            font-size: 3rem;
            margin-bottom: 1rem;
        }
        .title {
            font-size: 2rem;
            font-weight: 700;
            color: #1f2937;
            margin-bottom: 1rem;
        }
        .subtitle {
            font-size: 1.1rem;
            color: #6b7280;
            margin-bottom: 2rem;
        }
        .status {
            background: #f3f4f6;
            border-radius: 8px;
            padding: 1.5rem;
            margin: 2rem 0;
        }
        .spinner {
            border: 3px solid #f3f3f6;
            border-top: 3px solid #7c3aed;
            border-radius: 50%%;
            width: 40px;
            height: 40px;
            animation: spin 1s linear infinite;
            margin: 0 auto 1rem;
        }
        @keyframes spin {
            0%% { transform: rotate(0deg); }
            100%% { transform: rotate(360deg); }
        }
        .info {
            background: #eff6ff;
            border: 1px solid #dbeafe;
            border-radius: 8px;
            padding: 1rem;
            margin: 1rem 0;
            color: #1e40af;
        }
        .project-info {
            display: grid;
            grid-template-columns: auto 1fr;
            gap: 0.5rem 1rem;
            text-align: left;
            margin: 1.5rem 0;
        }
        .label {
            font-weight: 600;
            color: #374151;
        }
        .value {
            color: #6b7280;
            font-family: 'SF Mono', Monaco, monospace;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="logo">🚀</div>
        <h1 class="title">NSM Development Environment</h1>
        <p class="subtitle">Your development server is starting up...</p>
        
        <div class="status">
            <div class="spinner"></div>
            <p><strong>Starting %s project</strong></p>
            <p>Target: <code>%s</code></p>
        </div>
        
        <div class="project-info">
            <span class="label">Project:</span>
            <span class="value">%s</span>
            <span class="label">Type:</span>
            <span class="value">%s</span>
            <span class="label">Domain:</span>
            <span class="value">%s</span>
        </div>
        
        <div class="info">
            <strong>⏱️ This usually takes 10-30 seconds</strong><br>
            The page will automatically refresh once your server is ready.
        </div>
    </div>
    
    <script>
        // Auto-refresh every 2 seconds
        setTimeout(() => location.reload(), 2000);
        
        // Add some visual feedback
        let dots = 0;
        setInterval(() => {
            const status = document.querySelector('.status p');
            if (status) {
                dots = (dots + 1) %% 4;
                status.innerHTML = '<strong>Starting %s project' + '.'.repeat(dots) + '</strong>';
            }
        }, 500);
    </script>
</body>
</html>`,
		p.cfg.ProjectName,
		p.targetURL.String(),
		p.cfg.ProjectName,
		string(p.cfg.ProjectType),
		p.cfg.Domain,
		p.cfg.ProjectName)

	fmt.Fprint(w, html)
}

func (p *ProxyServer) getScheme() string {
	if p.cfg.EnableHTTPS {
		return "https"
	}
	return "http"
}

func (p *ProxyServer) getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.Split(xff, ",")[0]
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to remote address
	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}

	return r.RemoteAddr
}

func (p *ProxyServer) waitForReady(ctx context.Context, port int) error {
	timeout := time.NewTimer(10 * time.Second)
	defer timeout.Stop()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout.C:
			return fmt.Errorf("proxy server failed to start within timeout")
		case <-ticker.C:
			conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), time.Second)
			if err == nil {
				conn.Close()
				return nil
			}
		}
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
