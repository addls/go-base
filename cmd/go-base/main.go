package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed templates/api/*
var apiTemplateFS embed.FS

//go:embed templates/rpc/*
var rpcTemplateFS embed.FS

//go:embed templates/gateway/*
var gatewayTemplateFS embed.FS

//go:embed templates/rpc/register.tpl
var registerTemplateContent string

var (
	version = "v1.0.0"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "go-base",
		Short: "Go-Base CLI tool for initializing go-zero projects with go-base framework",
		Long: `Go-Base is a CLI tool that helps you initialize go-zero API projects
with the go-base enterprise framework base.

It automatically handles configuration file naming and integrates go-base templates.`,
		Version: version,
	}

	initCmd := &cobra.Command{
		Use:   "init [project-name]",
		Short: "Initialize a new go-zero project with go-base",
		Long: `Initialize a new go-zero business project with standard structure:
  - Creates project root directory with go.mod
  - Creates gateway directory with gateway service
  - Creates services/ping directory with RPC service

Examples:
  go-base init demo              # Initialize standard business project`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, args)
		},
	}

	upgradeCmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade go-base CLI tool to the latest patch version",
		Long: `Upgrade go-base CLI tool to the latest patch version within the current major version.

This command will:
  1. Upgrade the CLI tool to the latest patch version of the current major version
     (e.g., if current is v1.0.0, upgrade to v1.x.x latest)
  2. If run in a Go project directory, also upgrade the github.com/addls/go-base
     dependency to the same major version's latest patch version

This ensures CLI tool and project dependencies stay compatible.

Example:
  go-base upgrade`,
		RunE: runUpgrade,
	}

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(upgradeCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	// 检查项目目录是否已存在
	if _, err := os.Stat(projectName); err == nil {
		return fmt.Errorf("project directory '%s' already exists. Please remove it first or use a different name", projectName)
	}

	fmt.Printf("🚀 Initializing business project: %s\n", projectName)

	// 1. 检查并安装 goctl
	fmt.Println("\n📦 Step 1: Checking and installing goctl...")
	if err := checkAndInstallGoctl(); err != nil {
		return fmt.Errorf("failed to check/install goctl: %w", err)
	}
	fmt.Println("✓ goctl is ready")

	// 2. 安装公司级 goctl 模板（从嵌入的文件系统）
	fmt.Println("\n📋 Step 2: Installing go-base templates...")
	if err := installGoBaseTemplates(); err != nil {
		return fmt.Errorf("failed to install go-base templates: %w", err)
	}
	fmt.Println("✓ Templates installed")

	// 3. 创建项目主目录结构
	fmt.Println("\n🏗️  Step 3: Creating project structure...")
	if err := os.MkdirAll(projectName, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(projectName, "gateway"), 0755); err != nil {
		return fmt.Errorf("failed to create gateway directory: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(projectName, "services", "ping"), 0755); err != nil {
		return fmt.Errorf("failed to create services/ping directory: %w", err)
	}
	fmt.Println("✓ Project directories created")

	// 4. 在主目录下初始化 go.mod
	fmt.Println("\n📦 Step 4: Initializing go.mod...")
	modInitCmd := exec.Command("go", "mod", "init", projectName)
	modInitCmd.Dir = projectName
	modInitCmd.Stdout = os.Stdout
	modInitCmd.Stderr = os.Stderr
	if err := modInitCmd.Run(); err != nil {
		return fmt.Errorf("failed to init go.mod: %w", err)
	}
	fmt.Println("✓ go.mod initialized")

	// 5. 在 services/ping 目录下生成 proto 文件
	fmt.Println("\n📝 Step 5: Generating proto file in services/ping...")
	pingDir := filepath.Join(projectName, "services", "ping")
	rpcProtoCmd := exec.Command("goctl", "rpc", "-o", "ping.proto")
	rpcProtoCmd.Dir = pingDir
	rpcProtoCmd.Stdout = os.Stdout
	rpcProtoCmd.Stderr = os.Stderr
	if err := rpcProtoCmd.Run(); err != nil {
		return fmt.Errorf("failed to generate proto file: %w", err)
	}
	fmt.Println("✓ Proto file generated")

	// 6. 在 services/ping 目录下生成 RPC 服务代码
	fmt.Println("\n🔧 Step 6: Generating RPC service code...")
	protocCmd := exec.Command("goctl", "rpc", "protoc", "ping.proto",
		"--go_out=./pb",
		"--go-grpc_out=./pb",
		"--zrpc_out=.",
		"--client=true",
		"--style=go_zero",
		"-m")
	protocCmd.Dir = pingDir
	protocCmd.Stdout = os.Stdout
	protocCmd.Stderr = os.Stderr
	if err := protocCmd.Run(); err != nil {
		return fmt.Errorf("failed to generate RPC code: %w", err)
	}
	fmt.Println("✓ RPC service code generated")

	// 6.1 生成 internal/server/server.go 文件（使用 register.tpl）
	fmt.Println("\n📝 Step 6.1: Generating server registration file...")
	if err := generateServerRegisterFile(pingDir); err != nil {
		return fmt.Errorf("failed to generate server register file: %w", err)
	}
	fmt.Println("✓ Server registration file generated")

	// 6.2 确保 main.go 导入了 server 包
	fmt.Println("\n📝 Step 6.2: Updating main.go imports...")
	if err := ensureServerImportInMain(pingDir); err != nil {
		return fmt.Errorf("failed to update main.go imports: %w", err)
	}
	fmt.Println("✓ Main.go imports updated")

	// 6.3 重命名 RPC 服务配置文件为 config.yaml
	fmt.Println("\n📝 Step 6.3: Renaming RPC config file to config.yaml...")
	if err := renameRpcConfigFile(pingDir); err != nil {
		return fmt.Errorf("failed to rename RPC config file: %w", err)
	}
	fmt.Println("✓ RPC config file renamed to config.yaml")

	// 7. 在 gateway 目录下生成 Gateway 服务代码
	fmt.Println("\n🌐 Step 7: Generating Gateway service code...")
	gatewayDir := filepath.Join(projectName, "gateway")
	gatewayCmd := exec.Command("goctl", "gateway", "--dir", ".")
	gatewayCmd.Dir = gatewayDir
	gatewayCmd.Stdout = os.Stdout
	gatewayCmd.Stderr = os.Stderr
	if err := gatewayCmd.Run(); err != nil {
		return fmt.Errorf("failed to generate gateway code: %w", err)
	}
	fmt.Println("✓ Gateway service code generated")

	// 8. 重命名 gateway 配置文件
	gatewayConfigFile := filepath.Join(gatewayDir, "etc", "gateway.yaml")
	gatewayTargetFile := filepath.Join(gatewayDir, "etc", "config.yaml")
	if _, err := os.Stat(gatewayConfigFile); err == nil {
		if err := os.Rename(gatewayConfigFile, gatewayTargetFile); err != nil {
			fmt.Printf("⚠ Warning: failed to rename gateway config file: %v\n", err)
		} else {
			fmt.Println("✓ Gateway config file renamed to config.yaml")
		}
	}

	// 9. 生成 gateway 所需的 proto descriptor 文件
	fmt.Println("\n📝 Step 9: Generating proto descriptor file for gateway...")
	gatewayPbDir := filepath.Join(gatewayDir, "pb")
	if err := os.MkdirAll(gatewayPbDir, 0755); err != nil {
		return fmt.Errorf("failed to create gateway/pb directory: %w", err)
	}
	
	pingProtoFile := filepath.Join(projectName, "services", "ping", "ping.proto")
	
	// 检查 proto 文件是否存在
	if _, err := os.Stat(pingProtoFile); os.IsNotExist(err) {
		fmt.Printf("⚠ Warning: proto file not found: %s, skipping descriptor generation\n", pingProtoFile)
	} else {
		// 从项目根目录运行 protoc，使用相对路径
		protocCmd := exec.Command("protoc",
			"--descriptor_set_out", filepath.Join("gateway", "pb", "ping.pb"),
			"--include_imports",
			filepath.Join("services", "ping", "ping.proto"))
		protocCmd.Dir = projectName
		protocCmd.Stdout = os.Stdout
		protocCmd.Stderr = os.Stderr
		if err := protocCmd.Run(); err != nil {
			fmt.Printf("⚠ Warning: failed to generate proto descriptor file: %v\n", err)
			fmt.Printf("   You can manually run: protoc --descriptor_set_out=gateway/pb/ping.pb --include_imports services/ping/ping.proto\n")
		} else {
			fmt.Println("✓ Proto descriptor file generated: gateway/pb/ping.pb")
		}
	}

	// 10. 执行 go mod tidy
	fmt.Println("\n📦 Step 10: Running go mod tidy...")
	modCmd := exec.Command("go", "mod", "tidy")
	modCmd.Dir = projectName
	modCmd.Stdout = os.Stdout
	modCmd.Stderr = os.Stderr
	if err := modCmd.Run(); err != nil {
		fmt.Printf("⚠ Warning: go mod tidy failed: %v\n", err)
	} else {
		fmt.Println("✓ Dependencies updated")
	}

	fmt.Printf("\n✅ Business project %s initialized successfully!\n", projectName)
	fmt.Printf("\nProject structure:\n")
	fmt.Printf("  %s/\n", projectName)
	fmt.Printf("  ├── go.mod\n")
	fmt.Printf("  ├── gateway/          # Gateway service\n")
	fmt.Printf("  │   ├── etc/\n")
	fmt.Printf("  │   │   └── config.yaml\n")
	fmt.Printf("  │   └── gateway.go\n")
	fmt.Printf("  └── services/\n")
	fmt.Printf("      └── ping/        # Ping RPC service\n")
	fmt.Printf("          ├── ping.proto\n")
	fmt.Printf("          ├── etc/\n")
	fmt.Printf("          │   └── config.yaml\n")
	fmt.Printf("          └── ping.go\n")
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. cd %s\n", projectName)
	fmt.Printf("  2. Edit services/ping/ping.proto to define your RPC service\n")
	fmt.Printf("  3. Regenerate RPC code: cd services/ping && goctl rpc protoc ping.proto --go_out=./pb --go-grpc_out=./pb --zrpc_out=. --client=true --style=go_zero -m\n")
	fmt.Printf("  4. Edit gateway/etc/config.yaml to configure upstreams\n")
	fmt.Printf("  5. Run services: cd services/ping && go run ping.go\n")
	fmt.Printf("  6. Run gateway: cd gateway && go run gateway.go\n")

	return nil
}

// checkAndInstallGoctl 检查并安装 goctl
func checkAndInstallGoctl() error {
	// 先检查 goctl 是否已安装
	if _, err := exec.LookPath("goctl"); err == nil {
		// goctl 已安装，运行环境检查
		checkCmd := exec.Command("goctl", "env", "check", "--install", "--verbose", "--force")
		checkCmd.Stdout = os.Stdout
		checkCmd.Stderr = os.Stderr
		return checkCmd.Run()
	}

	// goctl 未安装，尝试安装
	fmt.Println("goctl not found, installing...")
	installCmd := exec.Command("go", "install", "github.com/zeromicro/go-zero/tools/goctl@latest")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install goctl: %w", err)
	}

	// 安装后运行环境检查
	checkCmd := exec.Command("goctl", "env", "check", "--install", "--verbose", "--force")
	checkCmd.Stdout = os.Stdout
	checkCmd.Stderr = os.Stderr
	return checkCmd.Run()
}

// installGoBaseTemplates 安装公司级 goctl 模板
func installGoBaseTemplates() error {
	// 1. 初始化 goctl 模板目录
	initCmd := exec.Command("goctl", "template", "init")
	initCmd.Stdout = os.Stdout
	initCmd.Stderr = os.Stderr
	if err := initCmd.Run(); err != nil {
		return fmt.Errorf("failed to init goctl templates: %w", err)
	}

	// 2. 获取 goctl 版本号
	versionCmd := exec.Command("goctl", "-v")
	versionOutput, err := versionCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get goctl version: %w", err)
	}

	// 解析版本号（格式：goctl version 1.8.5）
	versionStr := strings.TrimSpace(string(versionOutput))
	parts := strings.Fields(versionStr)
	var version string
	if len(parts) >= 3 {
		version = parts[2]
	} else {
		return fmt.Errorf("cannot parse goctl version from: %s", versionStr)
	}

	// 3. 复制 API 模板文件（从嵌入的文件系统）
	apiTemplateDir := filepath.Join(os.Getenv("HOME"), ".goctl", version, "api")
	if err := os.MkdirAll(apiTemplateDir, 0755); err != nil {
		return fmt.Errorf("failed to create api template directory: %w", err)
	}
	if err := copyTemplatesFromEmbed(apiTemplateFS, "templates/api", apiTemplateDir); err != nil {
		return fmt.Errorf("failed to copy api templates: %w", err)
	}

	// 4. 复制 RPC 模板文件（从嵌入的文件系统）
	rpcTemplateDir := filepath.Join(os.Getenv("HOME"), ".goctl", version, "rpc")
	if err := os.MkdirAll(rpcTemplateDir, 0755); err != nil {
		return fmt.Errorf("failed to create rpc template directory: %w", err)
	}
	if err := copyTemplatesFromEmbed(rpcTemplateFS, "templates/rpc", rpcTemplateDir); err != nil {
		return fmt.Errorf("failed to copy rpc templates: %w", err)
	}

	// 5. 复制 Gateway 模板文件（从嵌入的文件系统）
	gatewayTemplateDir := filepath.Join(os.Getenv("HOME"), ".goctl", version, "gateway")
	if err := os.MkdirAll(gatewayTemplateDir, 0755); err != nil {
		return fmt.Errorf("failed to create gateway template directory: %w", err)
	}
	if err := copyTemplatesFromEmbed(gatewayTemplateFS, "templates/gateway", gatewayTemplateDir); err != nil {
		return fmt.Errorf("failed to copy gateway templates: %w", err)
	}

	return nil
}

// copyTemplatesFromEmbed 从嵌入的文件系统复制模板文件
func copyTemplatesFromEmbed(embedFS embed.FS, srcDir, dstDir string) error {
	return fs.WalkDir(embedFS, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径（去掉 templates/api/ 前缀）
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dstDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		// 读取嵌入的文件
		data, err := embedFS.ReadFile(path)
		if err != nil {
			return err
		}

		// 写入目标文件
		return os.WriteFile(dstPath, data, 0644)
	})
}

// runUpgrade 执行升级命令
func runUpgrade(cmd *cobra.Command, args []string) error {
	fmt.Printf("🔄 Upgrading go-base CLI tool...\n")
	fmt.Printf("Current version: %s\n\n", version)

	// 检查 go 命令是否可用
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go command not found. Please install Go first: https://golang.org/dl/")
	}

	// 1. 升级 CLI 工具（升级到当前主版本的最新小版本）
	majorVersion := getMajorVersion(version)
	cliTarget := fmt.Sprintf("github.com/addls/go-base/cmd/go-base@%s", majorVersion)
	fmt.Printf("📦 Step 1: Upgrading go-base CLI tool to %s (latest patch version)...\n", majorVersion)
	installCmd := exec.Command("go", "install", cliTarget)
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr

	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to upgrade go-base CLI: %w\n\nPlease try manually: go install %s", err, cliTarget)
	}
	fmt.Println("✓ CLI tool upgraded")

	// 2. 检查当前目录是否是 Go 项目，如果是则升级依赖
	if err := upgradeProjectDependency(); err != nil {
		// 升级依赖失败不影响 CLI 工具升级，只打印警告
		majorVersion := getMajorVersion(version)
		fmt.Printf("\n⚠ Warning: Failed to upgrade project dependency: %v\n", err)
		fmt.Println("You can manually upgrade by running:")
		fmt.Printf("  go get github.com/addls/go-base@%s\n", majorVersion)
		fmt.Println("  go mod tidy")
	}

	fmt.Println("\n✅ Upgrade completed successfully!")
	fmt.Println("\nTo verify the new version, run:")
	fmt.Println("  go-base --version")

	return nil
}

// getMajorVersion 从版本号中提取主版本号（如 v1.0.0 -> v1）
func getMajorVersion(v string) string {
	// 移除前缀 "v" 如果存在
	v = strings.TrimPrefix(v, "v")

	// 按 "." 分割版本号
	parts := strings.Split(v, ".")
	if len(parts) > 0 {
		// 返回主版本号，如 "1" -> "v1"
		return "v" + parts[0]
	}

	// 如果无法解析，返回原版本号
	return v
}

// upgradeProjectDependency 升级当前项目中的 go-base 依赖
func upgradeProjectDependency() error {
	// 检查当前目录是否有 go.mod 文件
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		// 不是 Go 项目，跳过
		return nil
	}

	// 读取 go.mod 检查是否有 go-base 依赖
	goModData, err := os.ReadFile("go.mod")
	if err != nil {
		return fmt.Errorf("failed to read go.mod: %w", err)
	}

	// 检查是否包含 go-base 依赖
	if !strings.Contains(string(goModData), "github.com/addls/go-base") {
		// 没有 go-base 依赖，跳过
		return nil
	}

	// 从当前 CLI 版本中提取主版本号
	majorVersion := getMajorVersion(version)
	targetVersion := fmt.Sprintf("github.com/addls/go-base@%s", majorVersion)

	// 升级项目依赖
	fmt.Printf("\n📦 Step 2: Upgrading github.com/addls/go-base dependency to %s (latest patch version)...\n", majorVersion)

	// 使用 go get 更新依赖到当前主版本的最新小版本
	getCmd := exec.Command("go", "get", targetVersion)
	getCmd.Stdout = os.Stdout
	getCmd.Stderr = os.Stderr
	if err := getCmd.Run(); err != nil {
		return fmt.Errorf("failed to run go get: %w", err)
	}

	// 运行 go mod tidy 整理依赖
	fmt.Println("📦 Running go mod tidy...")
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Stdout = os.Stdout
	tidyCmd.Stderr = os.Stderr
	if err := tidyCmd.Run(); err != nil {
		return fmt.Errorf("failed to run go mod tidy: %w", err)
	}

	fmt.Println("✓ Project dependency upgraded")
	return nil
}

// generateServerRegisterFile 生成 internal/server/server.go 文件
// 直接遍历 internal/server 下的所有子目录，找到对应的 server 文件并生成注册代码
func generateServerRegisterFile(serviceDir string) error {
	modulePath := extractModulePath(serviceDir)
	if modulePath == "" {
		return fmt.Errorf("failed to extract module path from %s", serviceDir)
	}

	serverDir := filepath.Join(serviceDir, "internal", "server")
	pbDir := filepath.Join(serviceDir, "pb")

	// 检查 server 目录是否存在
	if _, err := os.Stat(serverDir); os.IsNotExist(err) {
		return fmt.Errorf("server directory does not exist: %s (make sure RPC code is generated first)", serverDir)
	}

	// 1. 遍历 internal/server 下的所有子目录
	serverSubDirs, err := os.ReadDir(serverDir)
	if err != nil {
		return fmt.Errorf("failed to read server directory: %w", err)
	}

	if len(serverSubDirs) == 0 {
		return fmt.Errorf("no server packages found in %s", serverDir)
	}

	var serviceRegistrations []string
	importMap := make(map[string]string) // import path -> alias

	for _, serverSubDir := range serverSubDirs {
		if !serverSubDir.IsDir() {
			continue
		}
		serverPkg := serverSubDir.Name()

		// 跳过 server.go 文件（如果存在）
		if serverPkg == "server.go" || strings.HasSuffix(serverPkg, ".go") {
			continue
		}

		// 2. 查找 server 文件，提取 NewXxxServer 函数
		serverFiles, err := filepath.Glob(filepath.Join(serverDir, serverPkg, "*_server.go"))
		if err != nil {
			fmt.Printf("⚠ Warning: failed to glob server files for %s: %v\n", serverPkg, err)
			continue
		}
		if len(serverFiles) == 0 {
			fmt.Printf("⚠ Warning: no server files found for package %s\n", serverPkg)
			continue
		}

		serverContent, err := os.ReadFile(serverFiles[0])
		if err != nil {
			continue
		}

		// 查找 NewXxxServer 函数
		serverLines := strings.Split(string(serverContent), "\n")
		var newServerFunc string
		for _, line := range serverLines {
			// 匹配格式：func NewXxxServer(...) *XxxServer
			if strings.Contains(line, "func New") && strings.Contains(line, "Server") {
				// 格式：func NewPingServer(svcCtx *svc.ServiceContext) *PingServer
				// 提取函数名：找到 "func " 后面的单词，直到遇到 "("
				funcIdx := strings.Index(line, "func ")
				if funcIdx >= 0 {
					funcPart := line[funcIdx+5:] // 跳过 "func "
					// 找到函数名的结束位置（空格或左括号）
					endIdx := strings.IndexAny(funcPart, " (")
					if endIdx > 0 {
						funcName := funcPart[:endIdx]
						if strings.HasPrefix(funcName, "New") && strings.HasSuffix(funcName, "Server") {
							newServerFunc = funcName
							break
						}
					}
				}
			}
		}

		if newServerFunc == "" {
			fmt.Printf("⚠ Warning: NewXxxServer function not found in server package %s\n", serverPkg)
			continue
		}

		// 3. 查找对应的 pb 包，提取 RegisterXxxServer 函数名
		// 优先查找同名的 pb 包
		pbPkg := serverPkg

		// 检查 pb 目录是否存在
		if _, err := os.Stat(pbDir); os.IsNotExist(err) {
			return fmt.Errorf("pb directory does not exist: %s (make sure RPC code is generated first)", pbDir)
		}

		grpcFiles, err := filepath.Glob(filepath.Join(pbDir, pbPkg, "*_grpc.pb.go"))
		if err != nil || len(grpcFiles) == 0 {
			// 如果同名包不存在，尝试查找所有 pb 包
			pbSubDirs, err := os.ReadDir(pbDir)
			if err != nil {
				fmt.Printf("⚠ Warning: failed to read pb directory: %v, skipping server package %s\n", err, serverPkg)
				continue
			}
			found := false
			for _, pbSubDir := range pbSubDirs {
				if !pbSubDir.IsDir() {
					continue
				}
				grpcFiles, err = filepath.Glob(filepath.Join(pbDir, pbSubDir.Name(), "*_grpc.pb.go"))
				if err == nil && len(grpcFiles) > 0 {
					pbPkg = pbSubDir.Name()
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("⚠ Warning: pb package not found for server package %s, skipping\n", serverPkg)
				continue
			}
		}

		grpcContent, err := os.ReadFile(grpcFiles[0])
		if err != nil {
			continue
		}

		// 查找 RegisterXxxServer 函数
		grpcLines := strings.Split(string(grpcContent), "\n")
		var registerFunc string
		for _, line := range grpcLines {
			// 匹配格式：func RegisterXxxServer(s grpc.ServiceRegistrar, srv XxxServer)
			if strings.Contains(line, "func Register") && strings.Contains(line, "Server") {
				// 提取函数名：找到 "func " 后面的单词，直到遇到 "("
				funcIdx := strings.Index(line, "func ")
				if funcIdx >= 0 {
					funcPart := line[funcIdx+5:] // 跳过 "func "
					// 找到函数名的结束位置（空格或左括号）
					endIdx := strings.IndexAny(funcPart, " (")
					if endIdx > 0 {
						funcName := funcPart[:endIdx]
						if strings.HasPrefix(funcName, "Register") && strings.HasSuffix(funcName, "Server") {
							registerFunc = funcName
							break
						}
					}
				}
			}
		}

		if registerFunc == "" {
			fmt.Printf("⚠ Warning: RegisterXxxServer function not found in pb package %s, skipping\n", pbPkg)
			continue
		}

		// 4. 生成注册代码
		// 格式：pbPkg.RegisterXxxServer(grpcServer, serverPkgAlias.NewXxxServer(ctx))
		pbImportPath := filepath.ToSlash(filepath.Join(modulePath, "pb", pbPkg))
		serverImportPath := filepath.ToSlash(filepath.Join(modulePath, "internal", "server", serverPkg))

		// 如果 pb 包名和 server 包名相同，需要为 server 包使用别名
		serverPkgAlias := serverPkg
		if pbPkg == serverPkg {
			// 使用 serverPkg 作为别名，例如：serverPing "test_project/services/ping/internal/server/ping"
			serverPkgAlias = "server" + strings.ToUpper(serverPkg[:1]) + serverPkg[1:]
		}

		registration := fmt.Sprintf("\t%s.%s(grpcServer, %s.%s(ctx))",
			pbPkg, registerFunc, serverPkgAlias, newServerFunc)
		serviceRegistrations = append(serviceRegistrations, registration)

		// 添加导入
		importMap[pbImportPath] = pbPkg
		// 如果包名相同，使用别名
		if pbPkg == serverPkg {
			importMap[serverImportPath] = serverPkgAlias
		} else {
			importMap[serverImportPath] = serverPkg
		}
	}

	if len(serviceRegistrations) == 0 {
		// 输出调试信息
		fmt.Printf("Debug: serverDir=%s, found %d server subdirs\n", serverDir, len(serverSubDirs))
		for _, subDir := range serverSubDirs {
			if subDir.IsDir() {
				fmt.Printf("Debug: found server subdir: %s\n", subDir.Name())
			}
		}
		return fmt.Errorf("no services found to register (checked %d server packages)", len(serverSubDirs))
	}

	// 5. 构建导入列表
	// 按照特定顺序排列：先 server 包（带别名），再 pb 包，最后 svc
	var importPackages []string
	var serverImports []string
	var pbImports []string

	for importPath, alias := range importMap {
		importLine := ""
		if alias == filepath.Base(importPath) {
			importLine = fmt.Sprintf("\t\"%s\"", importPath)
		} else {
			importLine = fmt.Sprintf("\t%s \"%s\"", alias, importPath)
		}

		// 分类：server 包（带别名）和 pb 包
		if strings.Contains(importPath, "internal/server") {
			serverImports = append(serverImports, importLine)
		} else if strings.Contains(importPath, "pb/") {
			pbImports = append(pbImports, importLine)
		} else {
			importPackages = append(importPackages, importLine)
		}
	}

	// 按顺序添加：先 server 包（带别名），再 pb 包
	importPackages = append(importPackages, serverImports...)
	importPackages = append(importPackages, pbImports...)

	// 最后添加 svc 导入
	svcImport := filepath.ToSlash(filepath.Join(modulePath, "internal", "svc"))
	importPackages = append(importPackages, fmt.Sprintf("\t\"%s\"", svcImport))

	// 6. 生成 server.go 文件（直接放在 internal/server/ 目录下）
	serverGoPath := filepath.Join(serverDir, "server.go")

	// 替换模板变量
	content := registerTemplateContent
	content = strings.ReplaceAll(content, "{{.importPackages}}", strings.Join(importPackages, "\n"))
	content = strings.ReplaceAll(content, "{{.serviceRegistrations}}", strings.Join(serviceRegistrations, "\n"))

	// 7. 写入文件
	if err := os.WriteFile(serverGoPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write server.go: %w", err)
	}

	return nil
}

// extractModulePath 从服务目录提取模块路径
func extractModulePath(serviceDir string) string {
	// 向上查找 go.mod 文件
	dir := serviceDir
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// 读取 go.mod 获取模块名
			content, err := os.ReadFile(goModPath)
			if err == nil {
				lines := strings.Split(string(content), "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "module ") {
						moduleName := strings.TrimSpace(strings.TrimPrefix(line, "module "))
						// 计算相对路径
						relPath, err := filepath.Rel(dir, serviceDir)
						if err == nil {
							// 使用 filepath.ToSlash 确保使用 / 作为路径分隔符（Go import 路径要求）
							return filepath.ToSlash(filepath.Join(moduleName, relPath))
						}
						return moduleName
					}
				}
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// ensureServerImportInMain 确保 main.go 文件导入了 server 包
// 使用 goimports 自动处理导入（添加缺失的导入，移除未使用的导入）
func ensureServerImportInMain(serviceDir string) error {
	mainGoPath := filepath.Join(serviceDir, filepath.Base(serviceDir)+".go")
	// 如果主文件不存在，尝试查找其他可能的文件名
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		// 尝试查找任何 .go 文件作为主文件
		goFiles, err := filepath.Glob(filepath.Join(serviceDir, "*.go"))
		if err != nil || len(goFiles) == 0 {
			return fmt.Errorf("main.go file not found in %s", serviceDir)
		}
		// 找到第一个非 test 的 go 文件
		for _, f := range goFiles {
			if !strings.HasSuffix(f, "_test.go") {
				mainGoPath = f
				break
			}
		}
	}

	// 检查 goimports 是否可用
	if _, err := exec.LookPath("goimports"); err != nil {
		// goimports 未安装，尝试安装
		fmt.Println("goimports not found, installing...")
		installCmd := exec.Command("go", "install", "golang.org/x/tools/cmd/goimports@latest")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("failed to install goimports: %w", err)
		}
	}

	// 先读取文件，检查是否需要添加 server 包的导入
	content, err := os.ReadFile(mainGoPath)
	if err != nil {
		return fmt.Errorf("failed to read main.go: %w", err)
	}

	contentStr := string(content)
	modulePath := extractModulePath(serviceDir)
	serverImportPath := filepath.ToSlash(filepath.Join(modulePath, "internal", "server"))

	// 检查是否已经导入了 server 包
	hasServerImport := false
	lines := strings.Split(contentStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, serverImportPath) {
			hasServerImport = true
			break
		}
	}

	// 如果没有导入 server 包，先添加它（临时添加，让 goimports 处理）
	if !hasServerImport {
		// 在 import 块中添加 server 导入
		importStart := strings.Index(contentStr, "import (")
		if importStart == -1 {
			return fmt.Errorf("cannot find import block in main.go (expected multi-line import)")
		}

		importEnd := strings.Index(contentStr[importStart:], ")")
		if importEnd == -1 {
			return fmt.Errorf("cannot find end of import block")
		}
		importEnd += importStart

		importBlock := contentStr[importStart : importEnd+1]
		newImport := fmt.Sprintf("\t\"%s\"\n", serverImportPath)

		// 在最后一个导入后添加
		lastQuoteIdx := strings.LastIndex(importBlock[:len(importBlock)-1], "\"")
		if lastQuoteIdx == -1 {
			return fmt.Errorf("cannot find last import in import block")
		}

		lastLineEnd := strings.LastIndex(importBlock[:lastQuoteIdx+1], "\n")
		if lastLineEnd == -1 {
			firstImport := strings.Index(importBlock, "\t")
			if firstImport == -1 {
				return fmt.Errorf("cannot find import statements in import block")
			}
			importBlock = importBlock[:firstImport] + newImport + importBlock[firstImport:]
		} else {
			importBlock = importBlock[:lastLineEnd+1] + newImport + importBlock[lastLineEnd+1:]
		}

		contentStr = contentStr[:importStart] + importBlock + contentStr[importEnd+1:]

		// 写入临时内容
		if err := os.WriteFile(mainGoPath, []byte(contentStr), 0644); err != nil {
			return fmt.Errorf("failed to write main.go: %w", err)
		}
	}

	// 使用 goimports 自动处理导入（添加缺失的，移除未使用的，格式化导入顺序）
	goimportsCmd := exec.Command("goimports", "-w", filepath.Base(mainGoPath))
	goimportsCmd.Dir = serviceDir
	goimportsCmd.Stdout = os.Stdout
	goimportsCmd.Stderr = os.Stderr
	if err := goimportsCmd.Run(); err != nil {
		return fmt.Errorf("failed to run goimports: %w", err)
	}

	return nil
}

// renameRpcConfigFile 重命名 RPC 服务的配置文件为 config.yaml
func renameRpcConfigFile(serviceDir string) error {
	etcDir := filepath.Join(serviceDir, "etc")

	// 查找 etc 目录下的所有 yaml 文件
	yamlFiles, err := filepath.Glob(filepath.Join(etcDir, "*.yaml"))
	if err != nil {
		return fmt.Errorf("failed to find yaml files: %w", err)
	}

	targetFile := filepath.Join(etcDir, "config.yaml")

	// 如果 config.yaml 已经存在，不需要重命名
	if _, err := os.Stat(targetFile); err == nil {
		return nil
	}

	// 查找需要重命名的文件（排除 config.yaml）
	var sourceFile string
	for _, yamlFile := range yamlFiles {
		if filepath.Base(yamlFile) != "config.yaml" {
			sourceFile = yamlFile
			break
		}
	}

	if sourceFile == "" {
		// 没有找到配置文件，可能 goctl 没有生成，或者已经重命名了
		return nil
	}

	// 重命名文件
	if err := os.Rename(sourceFile, targetFile); err != nil {
		return fmt.Errorf("failed to rename %s to %s: %w", sourceFile, targetFile, err)
	}

	return nil
}
