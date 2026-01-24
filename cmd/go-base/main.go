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

	// Check whether the project directory already exists.
	if _, err := os.Stat(projectName); err == nil {
		return fmt.Errorf("project directory '%s' already exists. Please remove it first or use a different name", projectName)
	}

	fmt.Printf("🚀 Initializing business project: %s\n", projectName)

	// 1. Check and install goctl.
	fmt.Println("\n📦 Step 1: Checking and installing goctl...")
	if err := checkAndInstallGoctl(); err != nil {
		return fmt.Errorf("failed to check/install goctl: %w", err)
	}
	fmt.Println("✓ goctl is ready")

	// 2. Install company-level goctl templates (from the embedded filesystem).
	fmt.Println("\n📋 Step 2: Installing go-base templates...")
	if err := installGoBaseTemplates(); err != nil {
		return fmt.Errorf("failed to install go-base templates: %w", err)
	}
	fmt.Println("✓ Templates installed")

	// 3. Create the project root directory structure.
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

	// 4. Initialize go.mod in the project root directory.
	fmt.Println("\n📦 Step 4: Initializing go.mod...")
	modInitCmd := exec.Command("go", "mod", "init", projectName)
	modInitCmd.Dir = projectName
	modInitCmd.Stdout = os.Stdout
	modInitCmd.Stderr = os.Stderr
	if err := modInitCmd.Run(); err != nil {
		return fmt.Errorf("failed to init go.mod: %w", err)
	}
	fmt.Println("✓ go.mod initialized")

	// 5. Generate proto file under services/ping.
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

	// 6. Generate RPC service code under services/ping.
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

	// 6.1 Generate internal/server/server.go (using register.tpl).
	fmt.Println("\n📝 Step 6.1: Generating server registration file...")
	if err := generateServerRegisterFile(pingDir); err != nil {
		return fmt.Errorf("failed to generate server register file: %w", err)
	}
	fmt.Println("✓ Server registration file generated")

	// 6.2 Ensure main.go imports the server package.
	fmt.Println("\n📝 Step 6.2: Updating main.go imports...")
	if err := ensureServerImportInMain(pingDir); err != nil {
		return fmt.Errorf("failed to update main.go imports: %w", err)
	}
	fmt.Println("✓ Main.go imports updated")

	// 6.3 Rename the RPC service config file to config.yaml.
	fmt.Println("\n📝 Step 6.3: Renaming RPC config file to config.yaml...")
	if err := renameRpcConfigFile(pingDir); err != nil {
		return fmt.Errorf("failed to rename RPC config file: %w", err)
	}
	fmt.Println("✓ RPC config file renamed to config.yaml")

	// 7. Generate Gateway service code under gateway.
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

	// 8. Rename the gateway config file.
	gatewayConfigFile := filepath.Join(gatewayDir, "etc", "gateway.yaml")
	gatewayTargetFile := filepath.Join(gatewayDir, "etc", "config.yaml")
	if _, err := os.Stat(gatewayConfigFile); err == nil {
		if err := os.Rename(gatewayConfigFile, gatewayTargetFile); err != nil {
			fmt.Printf("⚠ Warning: failed to rename gateway config file: %v\n", err)
		} else {
			fmt.Println("✓ Gateway config file renamed to config.yaml")
		}
	}

	// 9. Generate the proto descriptor file required by the gateway.
	fmt.Println("\n📝 Step 9: Generating proto descriptor file for gateway...")
	gatewayPbDir := filepath.Join(gatewayDir, "pb")
	if err := os.MkdirAll(gatewayPbDir, 0755); err != nil {
		return fmt.Errorf("failed to create gateway/pb directory: %w", err)
	}
	
	pingProtoFile := filepath.Join(projectName, "services", "ping", "ping.proto")
	
	// Check whether the proto file exists.
	if _, err := os.Stat(pingProtoFile); os.IsNotExist(err) {
		fmt.Printf("⚠ Warning: proto file not found: %s, skipping descriptor generation\n", pingProtoFile)
	} else {
		// Run protoc from the project root using relative paths.
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

	// 10. Run go mod tidy.
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

// checkAndInstallGoctl checks and installs goctl.
func checkAndInstallGoctl() error {
	// First, check if goctl is already installed.
	if _, err := exec.LookPath("goctl"); err == nil {
		// goctl is installed; run environment check.
		checkCmd := exec.Command("goctl", "env", "check", "--install", "--verbose", "--force")
		checkCmd.Stdout = os.Stdout
		checkCmd.Stderr = os.Stderr
		return checkCmd.Run()
	}

	// goctl is not installed; try installing it.
	fmt.Println("goctl not found, installing...")
	installCmd := exec.Command("go", "install", "github.com/zeromicro/go-zero/tools/goctl@latest")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install goctl: %w", err)
	}

	// Run environment check after installation.
	checkCmd := exec.Command("goctl", "env", "check", "--install", "--verbose", "--force")
	checkCmd.Stdout = os.Stdout
	checkCmd.Stderr = os.Stderr
	return checkCmd.Run()
}

// installGoBaseTemplates installs company-level goctl templates.
func installGoBaseTemplates() error {
	// 1. Initialize goctl template directory.
	initCmd := exec.Command("goctl", "template", "init")
	initCmd.Stdout = os.Stdout
	initCmd.Stderr = os.Stderr
	if err := initCmd.Run(); err != nil {
		return fmt.Errorf("failed to init goctl templates: %w", err)
	}

	// 2. Get goctl version.
	versionCmd := exec.Command("goctl", "-v")
	versionOutput, err := versionCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get goctl version: %w", err)
	}

	// Parse version string (format: goctl version 1.8.5).
	versionStr := strings.TrimSpace(string(versionOutput))
	parts := strings.Fields(versionStr)
	var version string
	if len(parts) >= 3 {
		version = parts[2]
	} else {
		return fmt.Errorf("cannot parse goctl version from: %s", versionStr)
	}

	// 3. Copy API template files (from embedded filesystem).
	apiTemplateDir := filepath.Join(os.Getenv("HOME"), ".goctl", version, "api")
	if err := os.MkdirAll(apiTemplateDir, 0755); err != nil {
		return fmt.Errorf("failed to create api template directory: %w", err)
	}
	if err := copyTemplatesFromEmbed(apiTemplateFS, "templates/api", apiTemplateDir); err != nil {
		return fmt.Errorf("failed to copy api templates: %w", err)
	}

	// 4. Copy RPC template files (from embedded filesystem).
	rpcTemplateDir := filepath.Join(os.Getenv("HOME"), ".goctl", version, "rpc")
	if err := os.MkdirAll(rpcTemplateDir, 0755); err != nil {
		return fmt.Errorf("failed to create rpc template directory: %w", err)
	}
	if err := copyTemplatesFromEmbed(rpcTemplateFS, "templates/rpc", rpcTemplateDir); err != nil {
		return fmt.Errorf("failed to copy rpc templates: %w", err)
	}

	// 5. Copy Gateway template files (from embedded filesystem).
	gatewayTemplateDir := filepath.Join(os.Getenv("HOME"), ".goctl", version, "gateway")
	if err := os.MkdirAll(gatewayTemplateDir, 0755); err != nil {
		return fmt.Errorf("failed to create gateway template directory: %w", err)
	}
	if err := copyTemplatesFromEmbed(gatewayTemplateFS, "templates/gateway", gatewayTemplateDir); err != nil {
		return fmt.Errorf("failed to copy gateway templates: %w", err)
	}

	return nil
}

// copyTemplatesFromEmbed copies template files from an embedded filesystem.
func copyTemplatesFromEmbed(embedFS embed.FS, srcDir, dstDir string) error {
	return fs.WalkDir(embedFS, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Compute relative path (strip templates/api/ prefix).
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dstDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		// Read the embedded file.
		data, err := embedFS.ReadFile(path)
		if err != nil {
			return err
		}

		// Write to destination file.
		return os.WriteFile(dstPath, data, 0644)
	})
}

// runUpgrade executes the upgrade command.
func runUpgrade(cmd *cobra.Command, args []string) error {
	fmt.Printf("🔄 Upgrading go-base CLI tool...\n")
	fmt.Printf("Current version: %s\n\n", version)

	// Check whether the go command is available.
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go command not found. Please install Go first: https://golang.org/dl/")
	}

	// 1. Upgrade CLI tool (to the latest patch within current major version).
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

	// 2. If current directory is a Go project, upgrade dependency as well.
	if err := upgradeProjectDependency(); err != nil {
		// Dependency upgrade failure does not affect CLI upgrade; only print a warning.
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

// getMajorVersion extracts the major version from a version string (e.g., v1.0.0 -> v1).
func getMajorVersion(v string) string {
	// Remove "v" prefix if present.
	v = strings.TrimPrefix(v, "v")

	// Split version by ".".
	parts := strings.Split(v, ".")
	if len(parts) > 0 {
		// Return major version, e.g. "1" -> "v1".
		return "v" + parts[0]
	}

	// If parsing fails, return original version string.
	return v
}

// upgradeProjectDependency upgrades github.com/addls/go-base dependency in the current project.
func upgradeProjectDependency() error {
	// Check whether go.mod exists in current directory.
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		// Not a Go project; skip.
		return nil
	}

	// Read go.mod to see whether go-base dependency is present.
	goModData, err := os.ReadFile("go.mod")
	if err != nil {
		return fmt.Errorf("failed to read go.mod: %w", err)
	}

	// Check whether go-base dependency is included.
	if !strings.Contains(string(goModData), "github.com/addls/go-base") {
		// No go-base dependency; skip.
		return nil
	}

	// Extract major version from current CLI version.
	majorVersion := getMajorVersion(version)
	targetVersion := fmt.Sprintf("github.com/addls/go-base@%s", majorVersion)

	// Upgrade project dependency.
	fmt.Printf("\n📦 Step 2: Upgrading github.com/addls/go-base dependency to %s (latest patch version)...\n", majorVersion)

	// Use go get to update dependency to the latest patch within current major version.
	getCmd := exec.Command("go", "get", targetVersion)
	getCmd.Stdout = os.Stdout
	getCmd.Stderr = os.Stderr
	if err := getCmd.Run(); err != nil {
		return fmt.Errorf("failed to run go get: %w", err)
	}

	// Run go mod tidy to tidy dependencies.
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

// generateServerRegisterFile generates internal/server/server.go.
// It walks all subdirectories under internal/server, finds server files, and generates registration code.
func generateServerRegisterFile(serviceDir string) error {
	modulePath := extractModulePath(serviceDir)
	if modulePath == "" {
		return fmt.Errorf("failed to extract module path from %s", serviceDir)
	}

	serverDir := filepath.Join(serviceDir, "internal", "server")
	pbDir := filepath.Join(serviceDir, "pb")

	// Check whether server directory exists.
	if _, err := os.Stat(serverDir); os.IsNotExist(err) {
		return fmt.Errorf("server directory does not exist: %s (make sure RPC code is generated first)", serverDir)
	}

	// 1. Walk all subdirectories under internal/server.
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

		// Skip server.go file (if present).
		if serverPkg == "server.go" || strings.HasSuffix(serverPkg, ".go") {
			continue
		}

		// 2. Find server file and extract the NewXxxServer function.
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

		// Find NewXxxServer function.
		serverLines := strings.Split(string(serverContent), "\n")
		var newServerFunc string
		for _, line := range serverLines {
			// Match: func NewXxxServer(...) *XxxServer
			if strings.Contains(line, "func New") && strings.Contains(line, "Server") {
				// Example: func NewPingServer(svcCtx *svc.ServiceContext) *PingServer
				// Extract function name: take the word after "func " until "(".
				funcIdx := strings.Index(line, "func ")
				if funcIdx >= 0 {
					funcPart := line[funcIdx+5:] // Skip "func "
					// Find the end of function name (space or left parenthesis).
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

		// 3. Find corresponding pb package and extract RegisterXxxServer function name.
		// Prefer a pb package with the same name.
		pbPkg := serverPkg

		// Check whether pb directory exists.
		if _, err := os.Stat(pbDir); os.IsNotExist(err) {
			return fmt.Errorf("pb directory does not exist: %s (make sure RPC code is generated first)", pbDir)
		}

		grpcFiles, err := filepath.Glob(filepath.Join(pbDir, pbPkg, "*_grpc.pb.go"))
		if err != nil || len(grpcFiles) == 0 {
			// If same-name package doesn't exist, try all pb packages.
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

		// Find RegisterXxxServer function.
		grpcLines := strings.Split(string(grpcContent), "\n")
		var registerFunc string
		for _, line := range grpcLines {
			// Match: func RegisterXxxServer(s grpc.ServiceRegistrar, srv XxxServer)
			if strings.Contains(line, "func Register") && strings.Contains(line, "Server") {
				// Extract function name: take the word after "func " until "(".
				funcIdx := strings.Index(line, "func ")
				if funcIdx >= 0 {
					funcPart := line[funcIdx+5:] // Skip "func "
					// Find the end of function name (space or left parenthesis).
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

		// 4. Generate registration code.
		// Format: pbPkg.RegisterXxxServer(grpcServer, serverPkgAlias.NewXxxServer(ctx))
		pbImportPath := filepath.ToSlash(filepath.Join(modulePath, "pb", pbPkg))
		serverImportPath := filepath.ToSlash(filepath.Join(modulePath, "internal", "server", serverPkg))

		// If pb package name is the same as server package name, use an alias for the server package.
		serverPkgAlias := serverPkg
		if pbPkg == serverPkg {
			// Use an alias, e.g. serverPing "test_project/services/ping/internal/server/ping"
			serverPkgAlias = "server" + strings.ToUpper(serverPkg[:1]) + serverPkg[1:]
		}

		registration := fmt.Sprintf("\t%s.%s(grpcServer, %s.%s(ctx))",
			pbPkg, registerFunc, serverPkgAlias, newServerFunc)
		serviceRegistrations = append(serviceRegistrations, registration)

		// Add imports.
		importMap[pbImportPath] = pbPkg
		// Use alias if package names are the same.
		if pbPkg == serverPkg {
			importMap[serverImportPath] = serverPkgAlias
		} else {
			importMap[serverImportPath] = serverPkg
		}
	}

	if len(serviceRegistrations) == 0 {
		// Output debug information.
		fmt.Printf("Debug: serverDir=%s, found %d server subdirs\n", serverDir, len(serverSubDirs))
		for _, subDir := range serverSubDirs {
			if subDir.IsDir() {
				fmt.Printf("Debug: found server subdir: %s\n", subDir.Name())
			}
		}
		return fmt.Errorf("no services found to register (checked %d server packages)", len(serverSubDirs))
	}

	// 5. Build import list.
	// Use a specific order: server packages (with alias) first, then pb packages, and finally svc.
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

		// Categorize: server packages (with alias) and pb packages.
		if strings.Contains(importPath, "internal/server") {
			serverImports = append(serverImports, importLine)
		} else if strings.Contains(importPath, "pb/") {
			pbImports = append(pbImports, importLine)
		} else {
			importPackages = append(importPackages, importLine)
		}
	}

	// Append in order: server packages (with alias) first, then pb packages.
	importPackages = append(importPackages, serverImports...)
	importPackages = append(importPackages, pbImports...)

	// Append svc import last.
	svcImport := filepath.ToSlash(filepath.Join(modulePath, "internal", "svc"))
	importPackages = append(importPackages, fmt.Sprintf("\t\"%s\"", svcImport))

	// 6. Generate server.go (directly under internal/server/).
	serverGoPath := filepath.Join(serverDir, "server.go")

	// Replace template variables.
	content := registerTemplateContent
	content = strings.ReplaceAll(content, "{{.importPackages}}", strings.Join(importPackages, "\n"))
	content = strings.ReplaceAll(content, "{{.serviceRegistrations}}", strings.Join(serviceRegistrations, "\n"))

	// 7. Write file.
	if err := os.WriteFile(serverGoPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write server.go: %w", err)
	}

	return nil
}

// extractModulePath extracts module path from a service directory.
func extractModulePath(serviceDir string) string {
	// Walk up to find go.mod.
	dir := serviceDir
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Read go.mod to get module name.
			content, err := os.ReadFile(goModPath)
			if err == nil {
				lines := strings.Split(string(content), "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "module ") {
						moduleName := strings.TrimSpace(strings.TrimPrefix(line, "module "))
						// Compute relative path.
						relPath, err := filepath.Rel(dir, serviceDir)
						if err == nil {
							// Use filepath.ToSlash to ensure "/" separators (required by Go import paths).
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

// ensureServerImportInMain ensures main.go imports the server package.
// It uses goimports to manage imports (add missing, remove unused) automatically.
func ensureServerImportInMain(serviceDir string) error {
	mainGoPath := filepath.Join(serviceDir, filepath.Base(serviceDir)+".go")
	// If main file doesn't exist, try other possible filenames.
	if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
		// Try any .go file as the main file.
		goFiles, err := filepath.Glob(filepath.Join(serviceDir, "*.go"))
		if err != nil || len(goFiles) == 0 {
			return fmt.Errorf("main.go file not found in %s", serviceDir)
		}
		// Find the first non-test go file.
		for _, f := range goFiles {
			if !strings.HasSuffix(f, "_test.go") {
				mainGoPath = f
				break
			}
		}
	}

	// Check whether goimports is available.
	if _, err := exec.LookPath("goimports"); err != nil {
		// goimports is not installed; try installing it.
		fmt.Println("goimports not found, installing...")
		installCmd := exec.Command("go", "install", "golang.org/x/tools/cmd/goimports@latest")
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr
		if err := installCmd.Run(); err != nil {
			return fmt.Errorf("failed to install goimports: %w", err)
		}
	}

	// Read file first and check whether we need to add server package import.
	content, err := os.ReadFile(mainGoPath)
	if err != nil {
		return fmt.Errorf("failed to read main.go: %w", err)
	}

	contentStr := string(content)
	modulePath := extractModulePath(serviceDir)
	serverImportPath := filepath.ToSlash(filepath.Join(modulePath, "internal", "server"))

	// Check whether the server package is already imported.
	hasServerImport := false
	lines := strings.Split(contentStr, "\n")
	for _, line := range lines {
		if strings.Contains(line, serverImportPath) {
			hasServerImport = true
			break
		}
	}

	// If server package is not imported, add it temporarily and let goimports handle it.
	if !hasServerImport {
		// Add server import into the import block.
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

		// Append after the last import.
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

		// Write temporary content.
		if err := os.WriteFile(mainGoPath, []byte(contentStr), 0644); err != nil {
			return fmt.Errorf("failed to write main.go: %w", err)
		}
	}

	// Use goimports to manage imports (add missing, remove unused, and format import order).
	goimportsCmd := exec.Command("goimports", "-w", filepath.Base(mainGoPath))
	goimportsCmd.Dir = serviceDir
	goimportsCmd.Stdout = os.Stdout
	goimportsCmd.Stderr = os.Stderr
	if err := goimportsCmd.Run(); err != nil {
		return fmt.Errorf("failed to run goimports: %w", err)
	}

	return nil
}

// renameRpcConfigFile renames the RPC service config file to config.yaml.
func renameRpcConfigFile(serviceDir string) error {
	etcDir := filepath.Join(serviceDir, "etc")

	// Find all yaml files under etc directory.
	yamlFiles, err := filepath.Glob(filepath.Join(etcDir, "*.yaml"))
	if err != nil {
		return fmt.Errorf("failed to find yaml files: %w", err)
	}

	targetFile := filepath.Join(etcDir, "config.yaml")

	// If config.yaml already exists, no need to rename.
	if _, err := os.Stat(targetFile); err == nil {
		return nil
	}

	// Find the file to rename (excluding config.yaml).
	var sourceFile string
	for _, yamlFile := range yamlFiles {
		if filepath.Base(yamlFile) != "config.yaml" {
			sourceFile = yamlFile
			break
		}
	}

	if sourceFile == "" {
		// No config file found: goctl may not have generated it, or it may already have been renamed.
		return nil
	}

	// Rename file.
	if err := os.Rename(sourceFile, targetFile); err != nil {
		return fmt.Errorf("failed to rename %s to %s: %w", sourceFile, targetFile, err)
	}

	return nil
}
