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

	var serviceType string
	initCmd := &cobra.Command{
		Use:   "init [project-name]",
		Short: "Initialize a new go-zero project with go-base",
		Long: `Initialize a new go-zero project (HTTP or gRPC) using goctl,
then automatically rename the config file to config.yaml.

Service types:
  http - HTTP/REST API service (default)
  rpc  - gRPC service

Examples:
  go-base init demo_project              # Initialize HTTP service
  go-base init demo_project --type http  # Initialize HTTP service
  go-base init demo_project --type rpc  # Initialize gRPC service`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, args, serviceType)
		},
	}
	initCmd.Flags().StringVarP(&serviceType, "type", "t", "http", "Service type: http or rpc")

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

func runInit(cmd *cobra.Command, args []string, serviceType string) error {
	projectName := args[0]

	// 验证服务类型
	serviceType = strings.ToLower(serviceType)
	if serviceType != "http" && serviceType != "rpc" {
		return fmt.Errorf("invalid service type: %s. Must be 'http' or 'rpc'", serviceType)
	}

	// 检查项目名称是否包含连字符（goctl 不支持）
	if strings.Contains(projectName, "-") {
		return fmt.Errorf("project name cannot contain hyphens (goctl limitation)")
	}

	// 检查项目目录是否已存在
	if _, err := os.Stat(projectName); err == nil {
		return fmt.Errorf("project directory '%s' already exists. Please remove it first or use a different name", projectName)
	}

	serviceTypeUpper := strings.ToUpper(serviceType)
	fmt.Printf("🚀 Initializing %s project: %s\n", serviceTypeUpper, projectName)

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

	// 3. 执行 goctl 命令创建项目结构
	fmt.Printf("\n🏗️  Step 3: Creating %s project structure...\n", serviceTypeUpper)
	var goctlCmd *exec.Cmd
	if serviceType == "http" {
		goctlCmd = exec.Command("goctl", "api", "new", projectName)
	} else {
		goctlCmd = exec.Command("goctl", "rpc", "new", projectName)
	}
	goctlCmd.Stdout = os.Stdout
	goctlCmd.Stderr = os.Stderr

	if err := goctlCmd.Run(); err != nil {
		return fmt.Errorf("failed to run goctl %s new: %w", serviceType, err)
	}

	// 重命名配置文件
	var configFile string
	if serviceType == "http" {
		// HTTP 服务：{project-name}-api.yaml -> config.yaml
		configFile = filepath.Join(projectName, "etc", projectName+"-api.yaml")
	} else {
		// RPC 服务：{project-name}.yaml -> config.yaml
		configFile = filepath.Join(projectName, "etc", projectName+".yaml")
	}
	targetFile := filepath.Join(projectName, "etc", "config.yaml")

	if _, err := os.Stat(configFile); err == nil {
		if err := os.Rename(configFile, targetFile); err != nil {
			return fmt.Errorf("failed to rename config file: %w", err)
		}
		fmt.Printf("✓ Renamed %s to %s\n", configFile, targetFile)
	} else {
		fmt.Printf("⚠ Config file %s not found, skipping rename\n", configFile)
	}

	// 4. 执行 go mod tidy
	fmt.Println("\n📦 Step 4: Running go mod tidy...")
	modCmd := exec.Command("go", "mod", "tidy")
	modCmd.Dir = projectName
	modCmd.Stdout = os.Stdout
	modCmd.Stderr = os.Stderr
	if err := modCmd.Run(); err != nil {
		fmt.Printf("⚠ Warning: go mod tidy failed: %v\n", err)
	} else {
		fmt.Println("✓ Dependencies updated")
	}

	// 5. 使用 goimports 清理未使用的导入
	fmt.Println("\n🧹 Step 5: Cleaning unused imports...")
	mainFile := projectName + ".go"
	mainFilePath := filepath.Join(projectName, mainFile)
	if _, err := os.Stat(mainFilePath); err == nil {
		// 检查 goimports 是否可用
		if _, err := exec.LookPath("goimports"); err == nil {
			importsCmd := exec.Command("goimports", "-w", mainFile)
			importsCmd.Dir = projectName
			importsCmd.Stdout = os.Stdout
			importsCmd.Stderr = os.Stderr
			if err := importsCmd.Run(); err != nil {
				fmt.Printf("⚠ Warning: goimports failed: %v\n", err)
			} else {
				fmt.Println("✓ Unused imports removed")
			}
		} else {
			// goimports 未安装，尝试安装
			fmt.Println("goimports not found, installing...")
			installCmd := exec.Command("go", "install", "golang.org/x/tools/cmd/goimports@latest")
			installCmd.Stdout = os.Stdout
			installCmd.Stderr = os.Stderr
			if err := installCmd.Run(); err == nil {
				// 安装成功后再次运行
				importsCmd := exec.Command("goimports", "-w", mainFile)
				importsCmd.Dir = projectName
				importsCmd.Stdout = os.Stdout
				importsCmd.Stderr = os.Stderr
				if err := importsCmd.Run(); err != nil {
					fmt.Printf("⚠ Warning: goimports failed: %v\n", err)
				} else {
					fmt.Println("✓ Unused imports removed")
				}
			} else {
				fmt.Printf("⚠ Warning: Failed to install goimports: %v\n", err)
				fmt.Println("You can manually install it: go install golang.org/x/tools/cmd/goimports@latest")
			}
		}
	}

	fmt.Printf("\n✅ %s project %s initialized successfully!\n", serviceTypeUpper, projectName)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. cd %s\n", projectName)
	if serviceType == "http" {
		fmt.Printf("  2. Edit api/%s.api to define your API\n", projectName)
		fmt.Printf("  3. Run: goctl api go -api api/%s.api -dir . -style go_zero\n", projectName)
		fmt.Printf("  4. Run: go run %s.go\n", projectName)
	} else {
		fmt.Printf("  2. Edit proto/%s.proto to define your gRPC service\n", projectName)
		fmt.Printf("  3. Run: goctl rpc protoc proto/%s.proto --go_out=. --go-grpc_out=. --zrpc_out=.\n", projectName)
		fmt.Printf("  4. Run: go run %s.go\n", projectName)
	}

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
