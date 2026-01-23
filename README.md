# Go-Base - 企业级 Go 框架底座

基于 go-zero 的企业级微服务框架底座，提供统一的启动方式、响应结构、错误码、中间件等基础能力。

## 快速开始

### 1. 安装 go-base CLI

```bash
# 从远程仓库安装
go install github.com/addls/go-base/cmd/go-base@latest

# 或本地开发安装
cd go-base
go install ./cmd/go-base

# 验证安装
go-base --version
```

### 2. 初始化项目

```bash
# 一键初始化标准业务项目（包含 gateway 和 services/ping）
go-base init demo
```

**生成的项目结构**：
```
demo/
├── go.mod
├── gateway/              # Gateway 服务
│   ├── etc/
│   │   └── config.yaml
│   ├── pb/              # proto descriptor 文件（自动生成）
│   │   └── ping.pb
│   └── main.go
└── services/
    └── ping/             # Ping RPC 服务
        ├── ping.proto    # Proto 定义
        ├── etc/
        │   └── config.yaml
        ├── pb/           # 生成的 proto 代码
        └── ping.go
```

**运行服务**：
```bash
# 1. 运行 RPC 服务
cd demo/services/ping
go run ping.go

# 2. 运行 Gateway 服务（在另一个终端）
cd demo/gateway
go run main.go
```

## 关键操作指南

### 创建 Proto 文件与生成代码

#### 1. 创建新的 RPC 服务

```bash
# 1. 进入项目根目录
cd demo

# 2. 创建服务目录
mkdir -p services/user
cd services/user

# 3. 生成 proto 模板文件
goctl rpc -o user.proto

# 4. 编辑 user.proto 定义服务接口
# syntax = "proto3";
# package user;
# option go_package="./user";
# 
# message GetUserReq {
#   int64 id = 1;
# }
# 
# message GetUserResp {
#   int64 id = 1;
#   string name = 2;
# }
# 
# service UserService {
#   rpc GetUser(GetUserReq) returns(GetUserResp);
# }

# 5. 生成 RPC 服务代码
goctl rpc protoc user.proto \
  --go_out=./pb \
  --go-grpc_out=./pb \
  --zrpc_out=. \
  --client=true \
  --style=go_zero \
  -m

# 6. 重命名配置文件（如果生成了其他名称）
mv etc/user.yaml etc/config.yaml 2>/dev/null || true

# 7. 运行服务
go run user.go
```

#### 2. 生成 Gateway 的 Proto Descriptor 文件（.pb）

**重要**：Gateway 需要 `.pb` descriptor 文件来解析 gRPC 服务定义，不能直接使用 `.proto` 源文件。

**自动生成**（推荐）：
- 使用 `go-base init` 初始化项目时，会自动生成 `gateway/pb/ping.pb`
- 如果后续添加了新服务，需要手动生成

**手动生成**：

```bash
# 在项目根目录执行
cd demo

# 生成 ping 服务的 descriptor 文件
protoc --descriptor_set_out=gateway/pb/ping.pb \
  --include_imports \
  services/ping/ping.proto

# 如果添加了新服务（如 user），也需要生成
protoc --descriptor_set_out=gateway/pb/user.pb \
  --include_imports \
  services/user/user.proto
```

**参数说明**：
- `--descriptor_set_out`：输出 descriptor 文件路径
- `--include_imports`：包含所有导入的 proto 文件（重要！Gateway 需要完整依赖）
- 最后一个参数：proto 源文件路径

### 配置 Gateway

在 `gateway/etc/config.yaml` 中配置路由：

**RPC 服务配置**：
```yaml
Upstreams:
  - Grpc:
      Target: localhost:8080  # gRPC 服务地址
    ProtoSets:
      - pb/ping.pb  # proto descriptor 文件路径（相对于 gateway 目录）
    Mappings:
      - Method: GET
        Path: /ping
        RpcPath: ping.Ping/Ping  # 格式：package.Service/Method
```

**RpcPath 格式说明**：
- 格式：`package.Service/Method`
- `package`：proto 文件中的 `package` 声明（如 `package ping;`）
- `Service`：服务名称（如 `service Ping {`）
- `Method`：RPC 方法名（如 `rpc Ping(...)`）

**示例**：
```protobuf
package user;
service UserService {
  rpc GetUser(...) returns (...);
}
```
对应的 RpcPath：`user.UserService/GetUser`

**HTTP 服务配置**：
```yaml
Upstreams:
  - Name: orderapi
    Http:
      Target: localhost:8888
      Prefix: /api
      Timeout: 3000
    Mappings:
      - Method: GET
        Path: /order
```

## 统一启动方式

### HTTP 服务

```go
package main

import (
    "github.com/zeromicro/go-zero/rest"
    "github.com/addls/go-base/pkg/bootstrap"
    "demo-project/internal/config"
    "demo-project/internal/handler"
    "demo-project/internal/svc"
)

func main() {
    bootstrap.RunHttp(bootstrap.WithHttpRoutes(func(server *rest.Server) {
        handler.RegisterHandlers(server, svc.NewServiceContext(*bootstrap.MustLoadConfig[config.Config]()))
    }))
}
```

### gRPC 服务

```go
package main

import (
    "github.com/addls/go-base/pkg/bootstrap"
    "google.golang.org/grpc"
    "demo-project/internal/config"
    "demo-project/internal/server"
    "demo-project/internal/svc"
)

func main() {
    bootstrap.RunRpc(bootstrap.WithRpcService(func(grpcServer *grpc.Server) {
        server.RegisterServices(grpcServer, svc.NewServiceContext(*bootstrap.MustLoadConfig[config.Config]()))
    }))
}
```

### Gateway 服务

```go
package main

import "github.com/addls/go-base/pkg/bootstrap"

func main() {
    // Gateway 配置通过配置文件中的 Upstreams 定义路由规则
    bootstrap.RunGateway()
}
```

## 配置说明

所有服务统一使用 `-f` flag 指定配置文件（默认 `etc/config.yaml`）：

```bash
# 使用默认配置
go run main.go

# 指定配置文件
go run main.go -f etc/config.prod.yaml
```

**配置结构**：
```go
// HTTP 服务配置
type Config struct {
    bootstrap.HttpConfig  // 嵌入 bootstrap.HttpConfig
    // 业务配置...
}

// gRPC 服务配置
type Config struct {
    bootstrap.RpcConfig  // 嵌入 bootstrap.RpcConfig
    // 业务配置...
}

// Gateway 服务配置
type Config struct {
    bootstrap.GatewayConfig  // 嵌入 bootstrap.GatewayConfig
    // Gateway 通过 Upstreams 配置路由
}
```

## 统一响应格式

```go
import "github.com/addls/go-base/pkg/response"

// 成功响应
response.Ok(w)                                    // 无数据
response.OkWithData(w, data)                     // 带数据
response.HandleResult(w, resp, err)              // 自动处理结果

// 错误响应
response.Error(w, err)                           // 自动识别错误码
response.ErrorWithCode(w, 20001, "自定义错误")    // 指定错误码
```

**响应格式**：
```json
{
  "code": 0,
  "msg": "success",
  "data": {},
  "traceId": "xxx"
}
```

## 错误码

```go
import "github.com/addls/go-base/pkg/errcode"

// 预定义错误码
errcode.ErrInvalidParam       // 20001 - 参数错误
errcode.ErrNotFound           // 20002 - 资源不存在
errcode.ErrUnauthorized       // 20004 - 未授权

// 自定义业务错误码
var ErrUserNotFound = errcode.New(30101, "用户不存在")
```

## 常见问题

### 1. Gateway 报错 "server does not support the reflection API"

**原因**：Gateway 需要 proto descriptor 文件（`.pb`），而不是 proto 源文件（`.proto`）

**解决**：
```bash
# 生成 descriptor 文件
protoc --descriptor_set_out=gateway/pb/ping.pb \
  --include_imports \
  services/ping/ping.proto
```

### 2. RpcPath 如何确定？

查看 proto 文件：
- `package ping;` → package 名是 `ping`
- `service Ping {` → 服务名是 `Ping`
- `rpc Ping(...)` → 方法名是 `Ping`

RpcPath = `ping.Ping/Ping`（格式：`package.Service/Method`）

### 3. ProtoSets 路径问题

`ProtoSets` 中的路径是相对于 `gateway` 目录的：
```yaml
ProtoSets:
  - pb/ping.pb  # 相对于 gateway 目录
```

## License

MIT License
