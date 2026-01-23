# {{.serviceName}} 配置文件

# ==================== 基础服务配置 (ServiceConf) ====================
# 服务名称，会出现在日志和追踪中
Name: {{.serviceName}}

# 服务运行模式: dev(开发), test(测试), rt(压测), pre(预发布), pro(生产)
# 默认值: pro
Mode: dev

# 指标上报 URL（可选，为空则禁用指标上报）
# MetricsUrl: http://localhost:9091/metrics

# ==================== gRPC 服务配置 (RpcServerConf) ====================
# 监听地址（格式：host:port 或 :port）
ListenOn: 0.0.0.0:50001

# 服务注册配置（可选，使用 etcd 进行服务发现）
# Etcd:
#   Hosts:
#     - localhost:2379
#   Key: {{.serviceName}}  # 服务注册键名
#   User: ""               # etcd 用户名（可选）
#   Pass: ""               # etcd 密码（可选）

# 是否启用认证（可选，默认 false）
# Auth: false

# Redis 配置（可选，启用认证时需要）
# Redis:
#   Host: localhost:6379
#   Type: node  # node(单节点), cluster(集群), sentinel(哨兵)
#   Pass: ""
#   DB: 0
#   # 集群模式配置
#   # Key: ""
#   # 哨兵模式配置
#   # Addrs:
#   #   - localhost:26379

# 是否启用严格控制（可选，默认 false）
# StrictControl: false

# 请求超时时间（毫秒），默认 2000 (2秒)
# Timeout: 2000

# CPU 阈值（0-1000），默认 900 (90%)，超过此阈值会触发限流
# CpuThreshold: 900

# 是否启用 gRPC 健康检查（可选，默认 true）
# Health: true

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

# ==================== gRPC 中间件配置 (ServerMiddlewaresConf) ====================
# 所有中间件默认启用，设置为 false 可禁用
Middlewares:
  # 链路追踪中间件
  Trace: true
  
  # 异常恢复中间件
  Recover: true
  
  # 统计中间件
  Stat: true
  
  # Prometheus 指标中间件
  Prometheus: true
  
  # 熔断器中间件
  Breaker: true

# ==================== Prometheus 配置 ====================
# Prometheus 监控配置（v1.4.3+ 已废弃，建议使用 MetricsUrl）
# Prometheus:
#   Host: 0.0.0.0
#   Port: 9091
#   Path: /metrics

# ==================== 分布式追踪配置 (Telemetry) ====================
# OpenTelemetry 追踪配置（可选）
# Telemetry:
#   Name: {{.serviceName}}           # 服务名称
#   Endpoint: http://localhost:4317  # 追踪数据上报地址
#   Sampler: 1.0                      # 采样率 (0.0-1.0)，默认 1.0
#   Batcher: otlpgrpc                 # 导出格式: zipkin, otlpgrpc, otlphttp, file
#   OtlpHeaders:                      # OTLP 自定义请求头
#     key: value
#   OtlpHttpPath: /v1/traces          # OTLP HTTP 路径
#   OtlpHttpSecure: false             # OTLP HTTP 是否启用 TLS
#   Disabled: false                   # 是否禁用追踪

# ==================== 开发服务器配置 (DevServer) ====================
# 开发服务器配置（v1.4.3+，可选）
# DevServer:
#   Port: 8848  # 开发服务器端口

# ==================== 应用配置 (go-base 扩展) ====================
# 应用配置
App:
  Name: {{.serviceName}}
  Version: 1.0.0
  Env: dev  # dev, test, prod

# ==================== 业务配置 ====================
# 数据库配置示例
# Database:
#   DataSource: "root:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
#   MaxOpenConns: 100
#   MaxIdleConns: 10
#   ConnMaxLifetime: 3600

# Redis 配置示例（如果未在认证中使用）
# Redis:
#   Host: localhost:6379
#   Pass: ""
#   DB: 0
#   PoolSize: 10
#   MinIdleConns: 5

# 其他业务配置...
# Custom:
#   Key: value
