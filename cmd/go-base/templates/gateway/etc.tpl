# Gateway config file

# ==================== Gateway base configuration ====================
# Gateway name
Name: Gateway

# Gateway listen host
Host: 0.0.0.0

# Gateway listen port
Port: 8888

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

# ==================== JWT configuration (go-base extension, optional) ====================
# JWT auth configuration (when enabled, Gateway verifies JWT before forwarding requests)
# After successful verification, Gateway passes user info through to backend gRPC
# (via grpc-gateway convention using the "Grpc-Metadata-" prefix headers):
# - Grpc-Metadata-x-jwt-user-id
# - Grpc-Metadata-x-jwt-user-name
# Jwt:
#   Secret: your-jwt-secret-key  # JWT secret (required)
#   SkipPaths:                    # Paths that skip JWT verification (optional)
#     - /ping
#     - /health

# ==================== Application configuration (go-base extension) ====================
# Application configuration
App:
  Name: Gateway
  Version: 1.0.0
  Env: dev  # dev, test, prod

# ==================== Gateway upstreams (Upstreams) ====================
# Gateway upstream service configuration
Upstreams:
  # HTTP-to-gRPC Gateway example
  - Grpc:
      Target: localhost:50001  # gRPC service address
      # Or use etcd service discovery
      # Etcd:
      #   Hosts:
      #     - localhost:2379
      #   Key: ping.rpc  # Service discovery key
    ProtoSets:
      - pb/ping.pb  # Proto descriptor path (relative to gateway directory)
    Mappings:
      - Method: GET
        Path: /ping
        RpcPath: ping.Ping/Ping
        # JWT config (optional; if global JWT is enabled, this documents how to skip per route)
        # Note: go-zero RouteMapping does not support custom fields; this is documentation only.
        # Actual control is via the global Jwt.SkipPaths config.
  
  # HTTP-to-HTTP Gateway example
  # - Name: userapi
  #   Http:
  #     Target: localhost:8080  # HTTP service address
  #     Prefix: /api  # Path prefix
  #     Timeout: 3000  # Timeout (milliseconds)
  #   Mappings:
  #     - Method: GET
  #       Path: /users
  #       # Forward to http://localhost:8080/api/users
