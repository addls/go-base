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
var templateFS embed.FS

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
		Short: "Initialize a new go-zero API project with go-base",
		Long: `Initialize a new go-zero API project using goctl api new,
then automatically rename the config file to config.yaml.

Example:
  go-base init demo_project`,
		Args: cobra.ExactArgs(1),
		RunE: runInit,
	}

	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	// 检查项目名称是否包含连字符（goctl 不支持）
	if strings.Contains(projectName, "-") {
		return fmt.Errorf("project name cannot contain hyphens (goctl limitation)")
	}

	fmt.Printf("🚀 Initializing project: %s\n", projectName)

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

	// 3. 执行 goctl api new
	fmt.Println("\n🏗️  Step 3: Creating project structure...")
	goctlCmd := exec.Command("goctl", "api", "new", projectName)
	goctlCmd.Stdout = os.Stdout
	goctlCmd.Stderr = os.Stderr

	if err := goctlCmd.Run(); err != nil {
		return fmt.Errorf("failed to run goctl api new: %w", err)
	}

	// 重命名配置文件
	configFile := filepath.Join(projectName, "etc", projectName+"-api.yaml")
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

	fmt.Printf("\n✅ Project %s initialized successfully!\n", projectName)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. cd %s\n", projectName)
	fmt.Printf("  2. Edit api/%s.api to define your API\n", projectName)
	fmt.Printf("  3. Run: goctl api go -api api/%s.api -dir . -style go_zero\n", projectName)
	fmt.Printf("  4. Run: go run %s.go\n", projectName)

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

	// 3. 复制模板文件（从嵌入的文件系统）
	goctlTemplateDir := filepath.Join(os.Getenv("HOME"), ".goctl", version, "api")
	if err := os.MkdirAll(goctlTemplateDir, 0755); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}

	// 从嵌入的文件系统复制模板文件
	return copyTemplatesFromEmbed(templateFS, "templates/api", goctlTemplateDir)
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
