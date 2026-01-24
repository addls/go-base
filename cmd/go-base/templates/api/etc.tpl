# {{.serviceName}} config file

# ==================== Base service configuration (ServiceConf) ====================
# Service name, appears in logs and tracing
Name: {{.serviceName}}

# Service mode: dev (development), test (testing), rt (load test), pre (pre-release), pro (production)
# Default: pro
Mode: dev

# Metrics push URL (optional; empty disables metrics push)
# MetricsUrl: http://localhost:9091/metrics

# ==================== HTTP service configuration (RestConf) ====================
# Listen host, default 0.0.0.0
Host: {{.host}}

# Listen port (required)
Port: {{.port}}

# HTTPS certificate config (optional; required when enabling HTTPS)
# CertFile: /path/to/cert.pem
# KeyFile: /path/to/key.pem

# Verbose logging (optional)
# Verbose: false

# Maximum concurrent connections, default 10000
# MaxConns: 10000

# Maximum request body size (bytes), default 1048576 (1MB)
# MaxBytes: 1048576

# Request timeout (milliseconds), default 3000 (3 seconds)
# Timeout: 3000

# CPU threshold (0-1000), default 900 (90%); exceeding triggers shedding
# CpuThreshold: 900

# ==================== Logging configuration (LogConf) ====================
Log:
  # Log mode: console, file, volume
  Mode: console
  
  # Log encoding: json (default), plain
  # Default: json
  # Encoding: json
  
  # Log level: debug, info, warn, error
  Level: info
  
  # Log file path (required for file/volume mode)
  # Path: logs
  
  # Compress log files (optional)
  # Compress: true
  
  # Log retention days (optional)
  # KeepDays: 7
  
  # Stack cooldown (milliseconds) to avoid frequent stack printing (optional)
  # StackCooldownMillis: 100

# ==================== Middleware configuration (MiddlewaresConf) ====================
# All middlewares are enabled by default; set to false to disable
Middlewares:
  # Tracing middleware
  Trace: true
  
  # Access log middleware
  Log: true
  
  # Prometheus metrics middleware
  Prometheus: true
  
  # Max connections limiting middleware
  MaxConns: true
  
  # Circuit breaker middleware
  Breaker: true
  
  # Shedding middleware (based on CPU threshold)
  Shedding: true
  
  # Timeout middleware
  Timeout: true
  
  # Recovery middleware
  Recover: true
  
  # Metrics collection middleware
  Metrics: true
  
  # Request body size limiting middleware
  MaxBytes: true
  
  # Gunzip middleware
  Gunzip: true

# ==================== Signature configuration (SignatureConf) ====================
# Signature verification config (optional)
# Signature:
#   Strict: false  # Strict mode
#   Expiry: 3600   # Signature expiry (seconds)

# ==================== Prometheus configuration ====================
# Prometheus monitoring config (deprecated since v1.4.3+; prefer MetricsUrl)
# Prometheus:
#   Host: 0.0.0.0
#   Port: 9091
#   Path: /metrics

# ==================== Distributed tracing (Telemetry) ====================
# OpenTelemetry tracing config (optional)
# Telemetry:
#   Name: {{.serviceName}}           # Service name
#   Endpoint: http://localhost:4317  # Trace exporter endpoint
#   Sampler: 1.0                      # Sampling rate (0.0-1.0), default 1.0
#   Batcher: otlpgrpc                 # Exporter: zipkin, otlpgrpc, otlphttp, file
#   OtlpHeaders:                      # Custom OTLP headers
#     key: value
#   OtlpHttpPath: /v1/traces          # OTLP HTTP path
#   OtlpHttpSecure: false             # Enable TLS for OTLP HTTP
#   Disabled: false                   # Disable tracing

# ==================== Dev server (DevServer) ====================
# Dev server config (v1.4.3+, optional)
# DevServer:
#   Port: 8848  # Dev server port

# ==================== Application configuration (go-base extension) ====================
# Application configuration
App:
  Name: {{.serviceName}}
  Version: 1.0.0
  Env: dev  # dev, test, prod

# ==================== Business configuration ====================
# Database configuration example
# Database:
#   DataSource: "root:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
#   MaxOpenConns: 100
#   MaxIdleConns: 10
#   ConnMaxLifetime: 3600

# Redis configuration example
# Redis:
#   Host: localhost:6379
#   Pass: ""
#   DB: 0
#   PoolSize: 10
#   MinIdleConns: 5

# Other business configuration...
# Custom:
#   Key: value
