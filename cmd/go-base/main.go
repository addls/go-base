package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
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

	// 执行 goctl api new
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

	fmt.Printf("✅ Project %s initialized successfully!\n", projectName)
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. cd %s\n", projectName)
	fmt.Printf("  2. Edit api/%s.api to define your API\n", projectName)
	fmt.Printf("  3. Run: goctl api go -api api/%s.api -dir . -style go_zero\n", projectName)
	fmt.Printf("  4. Run: go run %s.go\n", projectName)

	return nil
}
