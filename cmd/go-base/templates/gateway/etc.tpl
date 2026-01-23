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
