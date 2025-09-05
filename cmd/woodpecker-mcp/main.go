package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/denysvitali/woodpecker-ci-mcp/internal/auth"
	"github.com/denysvitali/woodpecker-ci-mcp/internal/client"
	"github.com/denysvitali/woodpecker-ci-mcp/internal/config"
	"github.com/denysvitali/woodpecker-ci-mcp/internal/server"
	"github.com/denysvitali/woodpecker-ci-mcp/tools"
)

var (
	cfgFile string
	logger  = logrus.New()

	// Styles
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFA500")).
		MarginBottom(1)

	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#10B981"))

	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EF4444"))

	infoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#0EA5E9"))
)

var rootCmd = &cobra.Command{
	Use:   "woodpecker-mcp",
	Short: "MCP Server for Woodpecker CI integration",
	Long: `Woodpecker MCP Server provides AI agents with access to Woodpecker CI
build statuses, pipeline management, and repository information through
the Model Context Protocol (MCP).

This server exposes various tools for:
- Listing and managing pipelines
- Viewing build statuses and logs
- Managing repositories
- Starting/stopping builds
- Approving pending pipelines`,
	Version:       "1.0.0",
	SilenceUsage:  true,
	SilenceErrors: true,
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MCP server",
	Long:  `Start the MCP server to handle requests from AI agents`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServer()
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `Manage Woodpecker MCP server configuration`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current configuration settings`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return showConfig()
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration",
	Long:  `Interactive setup of Woodpecker MCP server configuration`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return initConfig()
	},
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication management",
	Long:  `Manage authentication tokens for Woodpecker CI`,
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Woodpecker CI",
	Long:  `Interactively authenticate with your Woodpecker CI server`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return loginAuth()
	},
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Test connection to Woodpecker CI",
	Long:  `Test the connection to your configured Woodpecker CI server`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return testConnection()
	},
}

var testToolsCmd = &cobra.Command{
	Use:   "test-tools",
	Short: "Test all MCP tools",
	Long:  `Test all available MCP tools with interactive prompts or predefined test cases`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return testMCPTools()
	},
}

func init() {
	cobra.OnInitialize(initializeConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/woodpecker-mcp/config.yaml)")
	rootCmd.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("log-format", "text", "log format (text, json)")

	// Bind flags to viper
	if err := viper.BindPFlag("logging.level", rootCmd.PersistentFlags().Lookup("log-level")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind log-level flag")
	}
	if err := viper.BindPFlag("logging.format", rootCmd.PersistentFlags().Lookup("log-format")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind log-format flag")
	}

	// Add subcommands
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(testToolsCmd)

	// Config subcommands
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configInitCmd)

	// Auth subcommands
	authCmd.AddCommand(authLoginCmd)

	// Set default command to serve
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "serve")
	}
}

func initializeConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search for config in home directory
		configDir := filepath.Join(home, ".config", "woodpecker-mcp")
		viper.AddConfigPath(configDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	// Environment variables
	viper.SetEnvPrefix("WOODPECKER_MCP")
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if config file doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Debug().Err(err).Msg("Error reading config file")
		}
	}

	// Configure logger
	configureLogger()
}

func configureLogger() {
	// Set log level
	level := viper.GetString("logging.level")
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)

	// Set log format
	format := viper.GetString("logging.format")
	if format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}
}

func runServer() error {
	fmt.Println(titleStyle.Render("🚀 Starting Woodpecker MCP Server"))

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Configuration error: %v", err)))
		fmt.Println(infoStyle.Render("Run 'woodpecker-mcp config init' to set up configuration"))
		return err
	}

	// Create Woodpecker client
	wclient, err := client.New(client.Config{
		URL:   cfg.Woodpecker.URL,
		Token: cfg.Woodpecker.Token,
	}, logger)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Failed to connect to Woodpecker: %v", err)))
		return err
	}

	// Create MCP server
	mcpServer, err := server.NewMCPServer(cfg, wclient, logger)
	if err != nil {
		return fmt.Errorf("failed to create MCP server: %w", err)
	}

	fmt.Println(successStyle.Render("✅ MCP server started successfully"))
	fmt.Println(infoStyle.Render(fmt.Sprintf("Connected to Woodpecker server: %s", cfg.Woodpecker.URL)))

	// Start serving
	return mcpServer.Serve()
}

func showConfig() error {
	fmt.Println(titleStyle.Render("📋 Current Configuration"))

	cfg, err := config.Load()
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Failed to load configuration: %v", err)))
		return err
	}

	fmt.Printf("Server Name: %s\n", infoStyle.Render(cfg.Server.Name))
	fmt.Printf("Server Version: %s\n", infoStyle.Render(cfg.Server.Version))
	fmt.Printf("Woodpecker URL: %s\n", infoStyle.Render(cfg.Woodpecker.URL))

	if cfg.Woodpecker.Token != "" {
		fmt.Printf("Woodpecker Token: %s\n", infoStyle.Render(auth.MaskToken(cfg.Woodpecker.Token)))
	} else {
		fmt.Printf("Woodpecker Token: %s\n", errorStyle.Render("Not configured"))
	}

	fmt.Printf("Log Level: %s\n", infoStyle.Render(cfg.Logging.Level))
	fmt.Printf("Log Format: %s\n", infoStyle.Render(cfg.Logging.Format))

	return nil
}

func initConfig() error {
	fmt.Println(titleStyle.Render("🛠️  Woodpecker MCP Configuration Setup"))

	// Ensure config directory exists
	if err := config.EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Get Woodpecker URL
	url, err := auth.PromptForURL()
	if err != nil {
		return err
	}

	// Get authentication token
	token, err := auth.PromptForToken(url)
	if err != nil {
		return err
	}

	// Confirm configuration
	if !auth.ConfirmConfiguration(url, auth.MaskToken(token)) {
		fmt.Println(errorStyle.Render("Configuration cancelled"))
		return nil
	}

	// Create config file
	configDir, err := config.GetConfigDir()
	if err != nil {
		return err
	}

	configFile := filepath.Join(configDir, "config.yaml")

	configContent := fmt.Sprintf(`# Woodpecker MCP Server Configuration
server:
  name: "woodpecker-mcp"
  version: "1.0.0"

woodpecker:
  url: "%s"
  token: "%s"

logging:
  level: "info"
  format: "text"
`, url, token)

	err = os.WriteFile(configFile, []byte(configContent), 0600)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Println(successStyle.Render(fmt.Sprintf("✅ Configuration saved to: %s", configFile)))
	fmt.Println(infoStyle.Render("You can now run 'woodpecker-mcp serve' to start the server"))

	return nil
}

func loginAuth() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if cfg.Woodpecker.URL == "" {
		fmt.Println(errorStyle.Render("❌ Woodpecker URL not configured"))
		fmt.Println(infoStyle.Render("Run 'woodpecker-mcp config init' to set up configuration"))
		return nil
	}

	// Get new token
	token, err := auth.PromptForToken(cfg.Woodpecker.URL)
	if err != nil {
		return err
	}

	// Test the token
	wclient, err := client.New(client.Config{
		URL:   cfg.Woodpecker.URL,
		Token: token,
	}, logger)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Authentication failed: %v", err)))
		return err
	}

	// Get user info to verify token
	user, err := wclient.GetCurrentUser()
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Failed to verify authentication: %v", err)))
		return err
	}

	fmt.Println(successStyle.Render(fmt.Sprintf("✅ Successfully authenticated as: %s", user.Login)))

	// Update config file with new token
	viper.Set("woodpecker.token", token)

	// Write updated config
	configDir, err := config.GetConfigDir()
	if err != nil {
		return err
	}
	configFile := filepath.Join(configDir, "config.yaml")

	err = viper.WriteConfigAs(configFile)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Failed to save token: %v", err)))
		return err
	}

	fmt.Println(successStyle.Render("✅ Authentication token updated successfully"))

	return nil
}

func testConnection() error {
	fmt.Println(titleStyle.Render("🔍 Testing Woodpecker Connection"))

	cfg, err := config.Load()
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Failed to load configuration: %v", err)))
		return err
	}

	if err := cfg.Validate(); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Configuration error: %v", err)))
		return err
	}

	// Test connection
	wclient, err := client.New(client.Config{
		URL:   cfg.Woodpecker.URL,
		Token: cfg.Woodpecker.Token,
	}, logger)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Connection failed: %v", err)))
		return err
	}

	// Get user info
	user, err := wclient.GetCurrentUser()
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Failed to get user info: %v", err)))
		return err
	}

	// Get repositories count
	repos, err := wclient.ListRepositories()
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Failed to list repositories: %v", err)))
		return err
	}

	fmt.Println(successStyle.Render("✅ Connection successful!"))
	fmt.Printf("Server URL: %s\n", infoStyle.Render(cfg.Woodpecker.URL))
	fmt.Printf("Authenticated as: %s\n", infoStyle.Render(user.Login))
	fmt.Printf("Available repositories: %s\n", infoStyle.Render(fmt.Sprintf("%d", len(repos))))

	return nil
}

func testMCPTools() error {
	fmt.Println(titleStyle.Render("🧪 Testing MCP Tools"))

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Failed to load configuration: %v", err)))
		return err
	}

	if err := cfg.Validate(); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Configuration error: %v", err)))
		return err
	}

	// Create Woodpecker client
	wclient, err := client.New(client.Config{
		URL:   cfg.Woodpecker.URL,
		Token: cfg.Woodpecker.Token,
	}, logger)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Failed to connect to Woodpecker: %v", err)))
		return err
	}

	// Create tool manager
	tm := tools.NewToolManager(wclient, logger)

	// Interactive menu
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n" + titleStyle.Render("Available MCP Tools:"))
		fmt.Println("1. list_repositories - List all repositories")
		fmt.Println("2. get_repository - Get repository details")
		fmt.Println("3. list_pipelines - List pipelines for a repository")
		fmt.Println("4. get_pipeline_status - Get pipeline status")
		fmt.Println("5. start_pipeline - Start/restart a pipeline")
		fmt.Println("6. stop_pipeline - Stop a running pipeline")
		fmt.Println("7. approve_pipeline - Approve a pending pipeline")
		fmt.Println("8. get_logs - Get pipeline step logs")
		fmt.Println("9. test_all - Run all basic tests")
		fmt.Println("0. Exit")

		fmt.Print(infoStyle.Render("\nSelect a tool to test (0-9): "))

		if !scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			testListRepositories(tm)
		case "2":
			testGetRepository(tm, scanner)
		case "3":
			testListPipelines(tm, scanner)
		case "4":
			testGetPipelineStatus(tm, scanner)
		case "5":
			testStartPipeline(tm, scanner)
		case "6":
			testStopPipeline(tm, scanner)
		case "7":
			testApprovePipeline(tm, scanner)
		case "8":
			testGetLogs(tm, scanner)
		case "9":
			testAllTools(tm)
		case "0":
			fmt.Println(successStyle.Render("👋 Goodbye!"))
			return nil
		default:
			fmt.Println(errorStyle.Render("Invalid choice. Please select 0-9."))
		}
	}

	return nil
}

func testListRepositories(tm *tools.ToolManager) {
	fmt.Println(titleStyle.Render("Testing list_repositories"))

	ctx := context.Background()
	serverTools := tm.GetServerTools()

	// Find the list_repositories tool
	for _, serverTool := range serverTools {
		if serverTool.Tool.Name == "list_repositories" {
			// Test with all=false
			args := map[string]interface{}{"all": false}
			request := createMockCallToolRequest("list_repositories", args)

			fmt.Println(infoStyle.Render("Testing with all=false..."))
			result, err := serverTool.Handler(ctx, request)
			if err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Error: %v", err)))
				return
			}

			displayResult(result)

			// Test with all=true
			args = map[string]interface{}{"all": true}
			request = createMockCallToolRequest("list_repositories", args)

			fmt.Println(infoStyle.Render("Testing with all=true..."))
			result, err = serverTool.Handler(ctx, request)
			if err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Error: %v", err)))
				return
			}

			displayResult(result)
			break
		}
	}
}

func testGetRepository(tm *tools.ToolManager, scanner *bufio.Scanner) {
	fmt.Println(titleStyle.Render("Testing get_repository"))

	fmt.Print(infoStyle.Render("Enter repository name (owner/repo) or repository ID: "))
	if !scanner.Scan() {
		return
	}

	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		fmt.Println(errorStyle.Render("❌ Repository name/ID required"))
		return
	}

	ctx := context.Background()
	serverTools := tm.GetServerTools()

	for _, serverTool := range serverTools {
		if serverTool.Tool.Name == "get_repository" {
			var args map[string]interface{}

			// Check if input is numeric (ID) or string (name)
			if repoID, err := strconv.Atoi(input); err == nil {
				args = map[string]interface{}{"repo_id": repoID}
			} else {
				args = map[string]interface{}{"repo_name": input}
			}

			request := createMockCallToolRequest("get_repository", args)
			result, err := serverTool.Handler(ctx, request)
			if err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Error: %v", err)))
				return
			}

			displayResult(result)
			break
		}
	}
}

func testListPipelines(tm *tools.ToolManager, scanner *bufio.Scanner) {
	fmt.Println(titleStyle.Render("Testing list_pipelines"))

	fmt.Print(infoStyle.Render("Enter repository name (owner/repo) or repository ID: "))
	if !scanner.Scan() {
		return
	}

	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		fmt.Println(errorStyle.Render("❌ Repository name/ID required"))
		return
	}

	ctx := context.Background()
	serverTools := tm.GetServerTools()

	for _, serverTool := range serverTools {
		if serverTool.Tool.Name == "list_pipelines" {
			var args map[string]interface{}

			if repoID, err := strconv.Atoi(input); err == nil {
				args = map[string]interface{}{"repo_id": repoID}
			} else {
				args = map[string]interface{}{"repo_name": input}
			}

			request := createMockCallToolRequest("list_pipelines", args)
			result, err := serverTool.Handler(ctx, request)
			if err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Error: %v", err)))
				return
			}

			displayResult(result)
			break
		}
	}
}

func testGetPipelineStatus(tm *tools.ToolManager, scanner *bufio.Scanner) {
	fmt.Println(titleStyle.Render("Testing get_pipeline_status"))

	fmt.Print(infoStyle.Render("Enter repository name (owner/repo) or repository ID: "))
	if !scanner.Scan() {
		return
	}

	repoInput := strings.TrimSpace(scanner.Text())
	if repoInput == "" {
		fmt.Println(errorStyle.Render("❌ Repository name/ID required"))
		return
	}

	fmt.Print(infoStyle.Render("Enter pipeline number (or press Enter for latest): "))
	if !scanner.Scan() {
		return
	}

	pipelineInput := strings.TrimSpace(scanner.Text())

	ctx := context.Background()
	serverTools := tm.GetServerTools()

	for _, serverTool := range serverTools {
		if serverTool.Tool.Name == "get_pipeline_status" {
			args := make(map[string]interface{})

			if repoID, err := strconv.Atoi(repoInput); err == nil {
				args["repo_id"] = repoID
			} else {
				args["repo_name"] = repoInput
			}

			if pipelineInput == "" {
				args["latest"] = true
			} else if pipelineNum, err := strconv.Atoi(pipelineInput); err == nil {
				args["pipeline_number"] = pipelineNum
			} else {
				fmt.Println(errorStyle.Render("❌ Invalid pipeline number"))
				return
			}

			request := createMockCallToolRequest("get_pipeline_status", args)
			result, err := serverTool.Handler(ctx, request)
			if err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Error: %v", err)))
				return
			}

			displayResult(result)
			break
		}
	}
}

func testStartPipeline(tm *tools.ToolManager, scanner *bufio.Scanner) {
	fmt.Println(titleStyle.Render("Testing start_pipeline"))
	fmt.Println(errorStyle.Render("⚠️  This will actually start/restart a pipeline!"))

	fmt.Print(infoStyle.Render("Continue? (y/N): "))
	if !scanner.Scan() {
		return
	}

	if strings.ToLower(strings.TrimSpace(scanner.Text())) != "y" {
		fmt.Println(infoStyle.Render("Cancelled"))
		return
	}

	fmt.Print(infoStyle.Render("Enter repository name (owner/repo) or repository ID: "))
	if !scanner.Scan() {
		return
	}

	repoInput := strings.TrimSpace(scanner.Text())
	if repoInput == "" {
		fmt.Println(errorStyle.Render("❌ Repository name/ID required"))
		return
	}

	fmt.Print(infoStyle.Render("Enter pipeline number: "))
	if !scanner.Scan() {
		return
	}

	pipelineInput := strings.TrimSpace(scanner.Text())
	pipelineNum, err := strconv.Atoi(pipelineInput)
	if err != nil {
		fmt.Println(errorStyle.Render("❌ Invalid pipeline number"))
		return
	}

	ctx := context.Background()
	serverTools := tm.GetServerTools()

	for _, serverTool := range serverTools {
		if serverTool.Tool.Name == "start_pipeline" {
			args := map[string]interface{}{
				"pipeline_number": pipelineNum,
			}

			if repoID, err := strconv.Atoi(repoInput); err == nil {
				args["repo_id"] = repoID
			} else {
				args["repo_name"] = repoInput
			}

			request := createMockCallToolRequest("start_pipeline", args)
			result, err := serverTool.Handler(ctx, request)
			if err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Error: %v", err)))
				return
			}

			displayResult(result)
			break
		}
	}
}

func testStopPipeline(tm *tools.ToolManager, scanner *bufio.Scanner) {
	fmt.Println(titleStyle.Render("Testing stop_pipeline"))
	fmt.Println(errorStyle.Render("⚠️  This will actually stop a running pipeline!"))

	fmt.Print(infoStyle.Render("Continue? (y/N): "))
	if !scanner.Scan() {
		return
	}

	if strings.ToLower(strings.TrimSpace(scanner.Text())) != "y" {
		fmt.Println(infoStyle.Render("Cancelled"))
		return
	}

	fmt.Print(infoStyle.Render("Enter repository name (owner/repo) or repository ID: "))
	if !scanner.Scan() {
		return
	}

	repoInput := strings.TrimSpace(scanner.Text())
	if repoInput == "" {
		fmt.Println(errorStyle.Render("❌ Repository name/ID required"))
		return
	}

	fmt.Print(infoStyle.Render("Enter pipeline number: "))
	if !scanner.Scan() {
		return
	}

	pipelineInput := strings.TrimSpace(scanner.Text())
	pipelineNum, err := strconv.Atoi(pipelineInput)
	if err != nil {
		fmt.Println(errorStyle.Render("❌ Invalid pipeline number"))
		return
	}

	ctx := context.Background()
	serverTools := tm.GetServerTools()

	for _, serverTool := range serverTools {
		if serverTool.Tool.Name == "stop_pipeline" {
			args := map[string]interface{}{
				"pipeline_number": pipelineNum,
			}

			if repoID, err := strconv.Atoi(repoInput); err == nil {
				args["repo_id"] = repoID
			} else {
				args["repo_name"] = repoInput
			}

			request := createMockCallToolRequest("stop_pipeline", args)
			result, err := serverTool.Handler(ctx, request)
			if err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Error: %v", err)))
				return
			}

			displayResult(result)
			break
		}
	}
}

func testApprovePipeline(tm *tools.ToolManager, scanner *bufio.Scanner) {
	fmt.Println(titleStyle.Render("Testing approve_pipeline"))
	fmt.Println(errorStyle.Render("⚠️  This will actually approve a pending pipeline!"))

	fmt.Print(infoStyle.Render("Continue? (y/N): "))
	if !scanner.Scan() {
		return
	}

	if strings.ToLower(strings.TrimSpace(scanner.Text())) != "y" {
		fmt.Println(infoStyle.Render("Cancelled"))
		return
	}

	fmt.Print(infoStyle.Render("Enter repository name (owner/repo) or repository ID: "))
	if !scanner.Scan() {
		return
	}

	repoInput := strings.TrimSpace(scanner.Text())
	if repoInput == "" {
		fmt.Println(errorStyle.Render("❌ Repository name/ID required"))
		return
	}

	fmt.Print(infoStyle.Render("Enter pipeline number: "))
	if !scanner.Scan() {
		return
	}

	pipelineInput := strings.TrimSpace(scanner.Text())
	pipelineNum, err := strconv.Atoi(pipelineInput)
	if err != nil {
		fmt.Println(errorStyle.Render("❌ Invalid pipeline number"))
		return
	}

	ctx := context.Background()
	serverTools := tm.GetServerTools()

	for _, serverTool := range serverTools {
		if serverTool.Tool.Name == "approve_pipeline" {
			args := map[string]interface{}{
				"pipeline_number": pipelineNum,
			}

			if repoID, err := strconv.Atoi(repoInput); err == nil {
				args["repo_id"] = repoID
			} else {
				args["repo_name"] = repoInput
			}

			request := createMockCallToolRequest("approve_pipeline", args)
			result, err := serverTool.Handler(ctx, request)
			if err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Error: %v", err)))
				return
			}

			displayResult(result)
			break
		}
	}
}

func testGetLogs(tm *tools.ToolManager, scanner *bufio.Scanner) {
	fmt.Println(titleStyle.Render("Testing get_logs"))

	fmt.Print(infoStyle.Render("Enter repository name (owner/repo) or repository ID: "))
	if !scanner.Scan() {
		return
	}

	repoInput := strings.TrimSpace(scanner.Text())
	if repoInput == "" {
		fmt.Println(errorStyle.Render("❌ Repository name/ID required"))
		return
	}

	fmt.Print(infoStyle.Render("Enter pipeline number: "))
	if !scanner.Scan() {
		return
	}

	pipelineInput := strings.TrimSpace(scanner.Text())
	pipelineNum, err := strconv.Atoi(pipelineInput)
	if err != nil {
		fmt.Println(errorStyle.Render("❌ Invalid pipeline number"))
		return
	}

	fmt.Print(infoStyle.Render("Enter step ID: "))
	if !scanner.Scan() {
		return
	}

	stepInput := strings.TrimSpace(scanner.Text())
	stepID, err := strconv.Atoi(stepInput)
	if err != nil {
		fmt.Println(errorStyle.Render("❌ Invalid step ID"))
		return
	}

	fmt.Print(infoStyle.Render("Output format (json/text) [json]: "))
	if !scanner.Scan() {
		return
	}

	format := strings.TrimSpace(scanner.Text())
	if format == "" {
		format = "json"
	}

	ctx := context.Background()
	serverTools := tm.GetServerTools()

	for _, serverTool := range serverTools {
		if serverTool.Tool.Name == "get_logs" {
			args := map[string]interface{}{
				"pipeline_number": pipelineNum,
				"step_id":         stepID,
				"format":          format,
			}

			if repoID, err := strconv.Atoi(repoInput); err == nil {
				args["repo_id"] = repoID
			} else {
				args["repo_name"] = repoInput
			}

			request := createMockCallToolRequest("get_logs", args)
			result, err := serverTool.Handler(ctx, request)
			if err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("❌ Error: %v", err)))
				return
			}

			displayResult(result)
			break
		}
	}
}

func testAllTools(tm *tools.ToolManager) {
	fmt.Println(titleStyle.Render("Running basic tests for all tools"))

	// Test list_repositories (read-only)
	fmt.Println(infoStyle.Render("1. Testing list_repositories..."))
	testListRepositories(tm)

	fmt.Println(infoStyle.Render("✅ Basic tests completed. Interactive tests require specific repository information."))
}

func createMockCallToolRequest(toolName string, args map[string]interface{}) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		},
	}
}

func displayResult(result *mcp.CallToolResult) {
	if result.IsError {
		fmt.Println(errorStyle.Render("❌ Tool returned error:"))
	} else {
		fmt.Println(successStyle.Render("✅ Tool result:"))
	}

	for _, content := range result.Content {
		if textContent, ok := content.(mcp.TextContent); ok {
			// Try to pretty-print JSON
			var jsonData interface{}
			if json.Unmarshal([]byte(textContent.Text), &jsonData) == nil {
				if prettyJSON, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
					fmt.Println(string(prettyJSON))
					return
				}
			}
			// Fall back to plain text
			fmt.Println(textContent.Text)
		}
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
