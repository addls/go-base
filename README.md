# Go-Base - 企业级 Go 框架底座

基于 go-zero 的企业级微服务框架底座，提供统一的启动方式、响应结构、错误码、中间件等基础能力。

## 设计理念

- **统一入口**：所有服务类型（HTTP/gRPC/Gateway）使用统一的启动函数和配置方式
- **配置统一**：配置文件路径、加载方式、应用配置字段全部统一管理
- **命名规范**：所有类型和函数采用明确的前缀（Http/Rpc/Gateway）区分
- **职责分离**：启动逻辑、配置管理、响应处理、错误码等模块独立

## 特性

- **一句话启动**：`bootstrap.RunHttp()` 启动 HTTP 服务，`bootstrap.RunRpc()` 启动 gRPC 服务，`bootstrap.RunGateway()` 启动 Gateway 服务
- **统一响应结构**：标准化的 `{"code":0,"msg":"ok","data":{}}` 格式
- **统一错误码**：分层错误码体系，自动映射 HTTP 状态码
- **统一中间件**：链路追踪、访问日志、恢复、限流、认证等
- **Gateway 支持**：HTTP-to-HTTP 和 HTTP-to-gRPC 网关
- **goctl 公司模板**：代码生成统一风格，一键生成规范代码

## 快速开始

### 1. 安装 go-base CLI 工具

**方式一：从远程仓库安装（推荐）**

```bash
go install github.com/addls/go-base/cmd/go-base@latest
```

**方式二：本地开发安装**

```bash
# 克隆或进入 go-base 项目目录
cd go-base

# 本地安装
go install ./cmd/go-base

# 验证安装
go-base --version
```

### 2. 安装公司级 goctl 模板（可选）

**注意**：如果使用 `go-base init` 命令，模板会自动安装，此步骤可跳过。

**手动安装方式**（如果需要单独安装模板）：

```bash
# 1. 初始化 goctl 模板目录（只需执行一次）
goctl template init

# 2. 查找 goctl 版本号（例如：1.8.5）
goctl -v

# 3. 复制公司模板到对应版本的模板目录
# 注意：使用 /* 确保文件直接复制到 api 目录下，而不是多一层目录
cp -r cmd/go-base/templates/api/* ~/.goctl/$(goctl -v | awk '{print $3}')/api/

# 或者手动指定版本号（如果上面命令不工作）
# cp -r cmd/go-base/templates/api/* ~/.goctl/1.8.5/api/
```

**验证模板是否生效**：

```bash
# 检查 main.tpl 是否包含 go-base 相关内容
cat ~/.goctl/$(goctl -v | awk '{print $3}')/api/main.tpl | grep "github.com/addls/go-base"
# 应该看到：import "github.com/addls/go-base/pkg/bootstrap"
```

**开发说明**：模板文件已集成到 `go-base` 命令中（使用 Go embed），位于 `cmd/go-base/templates/api/` 目录。修改模板文件后，重新编译安装即可：

```bash
go install ./cmd/go-base
```

### 3. 创建业务项目

**推荐方式**：使用 go-base CLI 工具（全自动初始化）

```bash
# 使用 go-base init 初始化项目，会自动完成：
# 1. 检查并安装 goctl（如果未安装）
# 2. 安装公司级 goctl 模板
# 3. 创建项目结构
# 4. 将配置文件重命名为 config.yaml
# 5. 执行 go mod tidy
go-base init demo_project

# 编写 .api 文件后生成代码
cd demo_project
goctl api go -api api/demo_project.api -dir . -style go_zero
```

**或者**：手动使用 goctl

```bash
# 使用 goctl 生成项目（会自动使用已安装的公司模板）
goctl api new demo_project

# 编写 .api 文件后生成代码
goctl api go -api api/demo.api -dir . -style go_zero

# 注意：goctl 生成的配置文件默认名称是 {project-name}-api.yaml
# go-base 默认使用 etc/config.yaml，生成后需要重命名：
mv etc/demo_project-api.yaml etc/config.yaml

# 或者使用 -f 参数指定配置文件路径运行服务
```

### 4. 一句话启动

统一使用 `bootstrap.RunHttp()` 入口：

```go
package main

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/addls/github.com/addls/go-base/pkg/bootstrap"
	"demo-project/internal/config"
	"demo-project/internal/handler"
	"demo-project/internal/svc"
)

func main() {
	// -f 指定配置文件（默认 etc/config.yaml）
	bootstrap.RunHttp(bootstrap.WithHttpRoutes(func(server *rest.Server) {
		// 业务配置结构体嵌入 bootstrap.HttpConfig，确保字段统一
		ctx := svc.NewServiceContext(*bootstrap.MustLoadConfig[config.Config]())
		handler.RegisterHandlers(server, ctx)
	}))
}
```

## 目录结构

```
go-base/
├── go.mod
├── README.md
├── cmd/
│   └── go-base/       # CLI 工具
│       ├── main.go    # 命令入口
│       └── templates/  # 模板文件（嵌入到命令中）
│           └── api/    # goctl 公司级模板
├── pkg/
│   ├── bootstrap/     # 启动器（HTTP/gRPC/Gateway）
│   │   ├── http.go    # HTTP 服务启动
│   │   ├── rpc.go     # gRPC 服务启动
│   │   ├── gateway.go # Gateway 服务启动
│   │   └── context.go # 服务上下文
│   ├── config/        # 配置工具（AppConfig, ConfigFile）
│   ├── response/      # 统一响应
│   ├── errcode/       # 统一错误码
│   └── middleware/    # 统一中间件
```

## 模块说明

### bootstrap - 启动器

提供三种服务类型的统一启动入口，所有启动函数都通过 `-f` flag 统一控制配置文件路径（默认 `etc/config.yaml`）。

**HTTP 服务启动**：

```go
import (
    "github.com/zeromicro/go-zero/rest"
    "github.com/addls/github.com/addls/go-base/pkg/bootstrap"
    "github.com/addls/github.com/addls/go-base/pkg/config"
    "demo-project/internal/config"
    "demo-project/internal/handler"
    "demo-project/internal/svc"
)

func main() {
    bootstrap.RunHttp(bootstrap.WithHttpRoutes(func(server *rest.Server) {
        ctx := svc.NewServiceContext(*bootstrap.MustLoadConfig[config.Config]())
        handler.RegisterHandlers(server, ctx)
    }))
}
```

**gRPC 服务启动**：

```go
import (
    "google.golang.org/grpc"
    "github.com/addls/github.com/addls/go-base/pkg/bootstrap"
    "demo-project/internal/config"
    "demo-project/internal/svc"
    pb "demo-project/pb"
    server "demo-project/internal/server"
)

func main() {
    bootstrap.RunRpc(bootstrap.WithRpcService(func(grpcServer *grpc.Server) {
        ctx := svc.NewServiceContext(*bootstrap.MustLoadConfig[config.Config]())
        pb.RegisterYourServiceServer(grpcServer, server.NewYourServiceServer(ctx))
    }))
}
```

**Gateway 服务启动**：

```go
import "github.com/addls/go-base/pkg/bootstrap"

func main() {
    // Gateway 配置通过配置文件中的 Upstreams 定义路由规则
    bootstrap.RunGateway()
}
```

**启动选项（HttpOption）**：

```go
// 添加 HTTP 中间件
bootstrap.WithHttpMiddleware(middleware1, middleware2)

// 注册 HTTP 路由
bootstrap.WithHttpRoutes(func(server *rest.Server) { ... })

// HTTP 启动前/后回调
bootstrap.WithHttpBeforeStart(func(server *rest.Server) { ... })
bootstrap.WithHttpAfterStart(func(server *rest.Server) { ... })
```

**启动选项（RpcOption）**：

```go
// 添加 gRPC 拦截器
bootstrap.WithRpcInterceptor(interceptor1, interceptor2)

// 注册 gRPC 服务
bootstrap.WithRpcService(func(grpcServer *grpc.Server) { ... })

// gRPC 启动前/后回调
bootstrap.WithRpcBeforeStart(func(server *zrpc.RpcServer) { ... })
bootstrap.WithRpcAfterStart(func(server *zrpc.RpcServer) { ... })
```

**配置结构**：

```go
// HTTP 服务配置（用于 Gateway 或纯 HTTP 服务）
type Config struct {
    bootstrap.HttpConfig  // 嵌入 bootstrap.HttpConfig（包含 rest.RestConf + App）
    // 业务配置...
}

// gRPC 服务配置（用于 gRPC 服务）
type Config struct {
    bootstrap.RpcConfig  // 嵌入 bootstrap.RpcConfig（包含 zrpc.RpcServerConf + App）
    // 业务配置...
}

// Gateway 服务配置（用于 Gateway）
type Config struct {
    bootstrap.GatewayConfig  // 嵌入 bootstrap.GatewayConfig（包含 gateway.GatewayConf + App）
    // Gateway 通过 Upstreams 配置路由，通常不需要额外业务配置
}
```

**Gateway 配置示例**：

```yaml
# etc/gateway.yaml
Name: gateway
Host: 0.0.0.0
Port: 8888

# 应用配置
App:
  Name: gateway
  Version: 1.0.0
  Env: dev

# HTTP-to-gRPC Gateway
Upstreams:
  - Grpc:
      Target: localhost:50051
    ProtoSets:
      - hello.pb
    Mappings:
      - Method: GET
        Path: /ping
        RpcPath: hello.Hello/Ping

# HTTP-to-HTTP Gateway
# Upstreams:
#   - Name: userapi
#     Http:
#       Target: localhost:8080
#       Prefix: /api
#       Timeout: 3000
#     Mappings:
#       - Method: GET
#         Path: /users
```

**Gateway 模式（推荐）**：

go-zero 官方推荐使用 **Gateway 模式**（分离进程）：
- **gRPC 服务**：运行独立的 gRPC 服务器（使用 `bootstrap.RunRpc()`）
- **HTTP Gateway**：运行独立的 Gateway 服务（使用 `bootstrap.RunGateway()`），将 REST 请求转换为 gRPC 调用或转发到 HTTP 后端

优势：
- 关注点分离
- 独立扩展（HTTP 和 gRPC 可分别扩展）
- 独立部署和配置
- 更好的容错性（一个服务崩溃不影响另一个）

示例架构：
```
┌─────────────┐      HTTP/REST      ┌──────────────┐
│   Client    │ ──────────────────> │ HTTP Gateway │
└─────────────┘                      └──────────────┘
                                            │
                                            │ gRPC
                                            ▼
                                      ┌──────────────┐
                                      │ gRPC Server  │
                                      └──────────────┘
```

### config - 配置工具

```go
import "github.com/addls/github.com/addls/go-base/pkg/config"

// 获取配置文件路径（由 -f flag 控制）
config.ConfigFile() // 返回 "etc/config.yaml" 或命令行指定的路径

// AppConfig 应用配置结构
type AppConfig struct {
    Name    string // 应用名称
    Version string // 应用版本
    Env     string // 环境：dev, test, prod
}
```

### bootstrap - 配置加载

```go
import "github.com/addls/go-base/pkg/bootstrap"

// 加载任意配置结构体（从统一配置文件）
cfg := bootstrap.MustLoadConfig[YourConfig]()
```

### response - 统一响应

```go
import "github.com/addls/github.com/addls/go-base/pkg/response"

// 成功响应
response.Ok(w)
response.OkWithData(w, data)
response.OkWithPage(w, list, total, page, pageSize)

// 错误响应  
response.Error(w, err)
response.ErrorWithMsg(w, errcode.ErrInvalidParam, "name 不能为空")
```

### errcode - 错误码

```go
import "github.com/addls/github.com/addls/go-base/pkg/errcode"

// 预定义错误码
errcode.OK           // 0 - 成功
errcode.ErrInternal  // 10001 - 内部错误
errcode.ErrInvalidParam // 20001 - 参数错误

// 自定义错误码
var ErrUserNotFound = errcode.New(30101, "用户不存在")

// 带详情
err := errcode.ErrInvalidParam.WithMsg("name 字段不能为空")
```

## goctl 模板使用

### 安装模板

```bash
# 1. 初始化 goctl 模板目录（只需执行一次）
goctl template init

# 2. 查找 goctl 版本号（例如：1.8.5）
goctl -v

# 3. 复制公司模板到对应版本的模板目录
cp -r cmd/go-base/templates/api/* ~/.goctl/$(goctl -v | awk '{print $3}')/api/

# 或者手动指定版本号
# cp -r cmd/go-base/templates/api/* ~/.goctl/1.8.5/api/
```

### 生成代码

```bash
# 生成 API 服务
goctl api go -api api/user.api -dir . -style go_zero
```

**验证模板是否生效**：

```bash
# 检查 main.tpl 是否包含 go-base 相关内容
cat ~/.goctl/$(goctl -v | awk '{print $3}')/api/main.tpl | grep "go-base"
# 应该看到：import "github.com/addls/go-base/pkg/bootstrap" 和 "github.com/addls/go-base/pkg/config"
```

## License

MIT License
