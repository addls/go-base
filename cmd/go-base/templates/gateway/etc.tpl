# Gateway 配置文件

# ==================== Gateway 基础配置 ====================
# Gateway 名称
Name: Gateway

# Gateway 监听地址
Host: 0.0.0.0

# Gateway 监听端口
Port: 8888

# ==================== 日志配置 (LogConf) ====================
Log:
  # 日志模式: console(控制台), file(文件), volume(容器卷)
  Mode: console
  
  # 日志格式: json(JSON格式，默认), plain(文本格式)
  # 默认值: json
  # Encoding: json
  
  # 日志级别: debug, info, warn, error
  Level: info
  
  # 日志文件路径（file 或 volume 模式时需要）
  # Path: logs
  
  # 是否压缩日志文件（可选）
  # Compress: true
  
  # 日志保留天数（可选）
  # KeepDays: 7
  
  # 堆栈冷却时间（毫秒），用于避免频繁打印堆栈（可选）
  # StackCooldownMillis: 100

# ==================== 中间件配置 (MiddlewaresConf) ====================
# 所有中间件默认启用，设置为 false 可禁用
Middlewares:
  # 链路追踪中间件
  Trace: true
  
  # 访问日志中间件
  Log: true
  
  # Prometheus 指标中间件
  Prometheus: true
  
  # 最大连接数限制中间件
  MaxConns: true
  
  # 熔断器中间件
  Breaker: true
  
  # 限流中间件（基于 CPU 阈值）
  Shedding: true
  
  # 超时控制中间件
  Timeout: true
  
  # 异常恢复中间件
  Recover: true
  
  # 指标收集中间件
  Metrics: true
  
  # 请求体大小限制中间件
  MaxBytes: true
  
  # Gzip 解压缩中间件
  Gunzip: true

# ==================== JWT 配置 (go-base 扩展，可选) ====================
# JWT 认证配置（如果启用，Gateway 会在请求转发前进行 JWT 验证）
# 验证成功后，JWT claims 会通过 HTTP Header (X-Jwt-Claims) 透传给后端服务
# Jwt:
#   Secret: your-jwt-secret-key  # JWT 密钥（必需）
#   SkipPaths:                    # 跳过 JWT 验证的路径列表（可选）
#     - /ping
#     - /health

# ==================== 应用配置 (go-base 扩展) ====================
# 应用配置
App:
  Name: Gateway
  Version: 1.0.0
  Env: dev  # dev, test, prod

# ==================== Gateway 上游配置 (Upstreams) ====================
# Gateway 上游服务配置
Upstreams:
  # HTTP-to-gRPC Gateway 示例
  - Grpc:
      Target: localhost:50001  # gRPC 服务地址
      # 或者使用 etcd 服务发现
      # Etcd:
      #   Hosts:
      #     - localhost:2379
      #   Key: ping.rpc  # 服务发现键名
    ProtoSets:
      - pb/ping.pb  # proto 文件路径（相对于 gateway 目录）
    Mappings:
      - Method: GET
        Path: /ping
        RpcPath: ping.Ping/Ping
        # JWT 配置（可选，如果配置了全局 JWT，这里可以控制单个路由是否跳过）
        # 注意：go-zero 的 RouteMapping 不支持自定义字段，这里只是注释说明
        # 实际控制通过全局 Jwt.SkipPaths 配置
  
  # HTTP-to-HTTP Gateway 示例
  # - Name: userapi
  #   Http:
  #     Target: localhost:8080  # HTTP 服务地址
  #     Prefix: /api  # 路径前缀
  #     Timeout: 3000  # 超时时间（毫秒）
  #   Mappings:
  #     - Method: GET
  #       Path: /users
  #       # 转发到 http://localhost:8080/api/users
