use axum::{
    extract::Query,
    http::{HeaderMap, StatusCode},
    response::{Html, Json},
    routing::{get, post},
    Router,
};
use serde::{Deserialize, Serialize};
use std::{collections::HashMap, fs, net::SocketAddr};
use tower_http::{cors::CorsLayer, services::ServeDir};
use tracing::{info, warn};

#[derive(Serialize, Deserialize, Debug)]
struct NSMConfig {
    http: u16,
    https: u16,
    host: String,
}

impl Default for NSMConfig {
    fn default() -> Self {
        Self {
            http: {{.Port}},
            https: {{.HTTPSPort}},
            host: "127.0.0.1".to_string(),
        }
    }
}

#[derive(Serialize)]
struct AppInfo {
    name: String,
    version: String,
    domain: String,
    nsm_enabled: bool,
    timestamp: chrono::DateTime<chrono::Utc>,
    headers: Option<HashMap<String, String>>,
}

#[derive(Serialize)]
struct HealthResponse {
    status: String,
    timestamp: chrono::DateTime<chrono::Utc>,
    uptime: String,
}

fn load_nsm_config() -> NSMConfig {
    match fs::read_to_string(".nsm-ports.json") {
        Ok(contents) => match serde_json::from_str(&contents) {
            Ok(config) => {
                info!("🔧 NSM: Using HTTP port {}", config.http);
                config
            }
            Err(e) => {
                warn!("NSM: Failed to parse configuration: {}", e);
                NSMConfig::default()
            }
        },
        Err(_) => NSMConfig::default(),
    }
}

async fn home_handler() -> Html<&'static str> {
    Html(include_str!("../templates/index.html"))
}

async fn api_info_handler(
    Query(params): Query<HashMap<String, String>>,
    headers: HeaderMap,
) -> Json<AppInfo> {
    let mut header_map = HashMap::new();
    
    // Include debug headers if requested
    if params.get("debug").is_some() {
        for (name, value) in headers.iter() {
            if let Ok(value_str) = value.to_str() {
                header_map.insert(name.to_string(), value_str.to_string());
            }
        }
    }

    let nsm_enabled = std::env::var("NSM_ENABLED").unwrap_or_default() == "true";

    Json(AppInfo {
        name: "{{.ProjectName}}".to_string(),
        version: "1.0.0".to_string(),
        domain: "{{.Domain}}".to_string(),
        nsm_enabled,
        timestamp: chrono::Utc::now(),
        headers: if header_map.is_empty() { None } else { Some(header_map) },
    })
}

async fn health_handler() -> Json<HealthResponse> {
    Json(HealthResponse {
        status: "healthy".to_string(),
        timestamp: chrono::Utc::now(),
        uptime: "running".to_string(),
    })
}

#[derive(Deserialize)]
struct EchoRequest {
    message: String,
}

#[derive(Serialize)]
struct EchoResponse {
    echo: String,
    timestamp: chrono::DateTime<chrono::Utc>,
    id: String,
}

async fn echo_handler(Json(payload): Json<EchoRequest>) -> Json<EchoResponse> {
    Json(EchoResponse {
        echo: payload.message,
        timestamp: chrono::Utc::now(),
        id: uuid::Uuid::new_v4().to_string(),
    })
}

async fn not_found() -> (StatusCode, Json<serde_json::Value>) {
    (
        StatusCode::NOT_FOUND,
        Json(serde_json::json!({
            "error": "Not Found",
            "message": "The requested resource was not found",
            "timestamp": chrono::Utc::now()
        })),
    )
}

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    // Initialize tracing
    tracing_subscriber::fmt()
        .with_env_filter(
            std::env::var("RUST_LOG")
                .unwrap_or_else(|_| "{{.ProjectName | replace "_" "-"}}=debug,tower_http=debug".into()),
        )
        .init();

    let config = load_nsm_config();

    // Build our application with routes
    let app = Router::new()
        .route("/", get(home_handler))
        .route("/api/info", get(api_info_handler))
        .route("/api/health", get(health_handler))
        .route("/api/echo", post(echo_handler))
        .nest_service("/static", ServeDir::new("static"))
        .layer(CorsLayer::permissive())
        .fallback(not_found);

    let addr: SocketAddr = format!("{}:{}", config.host, config.http).parse()?;

    info!("🚀 Rust server starting on {}", addr);
    info!("🌐 Domain: {{.Domain}}");
    info!("📡 NSM: {}", if std::env::var("NSM_ENABLED").unwrap_or_default() == "true" { "Enabled" } else { "Disabled" });
    info!("🦀 Framework: Axum");
    println!();

    let listener = tokio::net::TcpListener::bind(addr).await?;
    axum::serve(listener, app).await?;

    Ok(())
}
