package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/denysvitali/woodpecker-ci-mcp/internal/client"
	"github.com/denysvitali/woodpecker-ci-mcp/internal/config"
	"github.com/denysvitali/woodpecker-ci-mcp/tools"
)

func testMCPTools() error {
	fmt.Println(titleStyle.Render("Testing MCP Tools"))

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Failed to load configuration: %v", err)))
		return err
	}

	if err := cfg.Validate(); err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Configuration error: %v", err)))
		return err
	}

	// Create Woodpecker client
	wclient, err := client.New(client.Config{
		URL:   cfg.Woodpecker.URL,
		Token: cfg.Woodpecker.Token,
	}, logger)
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("Failed to connect to Woodpecker: %v", err)))
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
			fmt.Println(successStyle.Render("Goodbye!"))
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
				fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err)))
				return
			}

			displayResult(result)

			// Test with all=true
			args = map[string]interface{}{"all": true}
			request = createMockCallToolRequest("list_repositories", args)

			fmt.Println(infoStyle.Render("Testing with all=true..."))
			result, err = serverTool.Handler(ctx, request)
			if err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err)))
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
		fmt.Println(errorStyle.Render("Repository name/ID required"))
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
				fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err)))
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
		fmt.Println(errorStyle.Render("Repository name/ID required"))
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
				fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err)))
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
		fmt.Println(errorStyle.Render("Repository name/ID required"))
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
				fmt.Println(errorStyle.Render("Invalid pipeline number"))
				return
			}

			request := createMockCallToolRequest("get_pipeline_status", args)
			result, err := serverTool.Handler(ctx, request)
			if err != nil {
				fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err)))
				return
			}

			displayResult(result)
			break
		}
	}
}

func testStartPipeline(tm *tools.ToolManager, scanner *bufio.Scanner) {
	fmt.Println(titleStyle.Render("Testing start_pipeline"))
	fmt.Println(errorStyle.Render("This will actually start/restart a pipeline!"))

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
		fmt.Println(errorStyle.Render("Repository name/ID required"))
		return
	}

	fmt.Print(infoStyle.Render("Enter pipeline number: "))
	if !scanner.Scan() {
		return
	}

	pipelineInput := strings.TrimSpace(scanner.Text())
	pipelineNum, err := strconv.Atoi(pipelineInput)
	if err != nil {
		fmt.Println(errorStyle.Render("Invalid pipeline number"))
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
				fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err)))
				return
			}

			displayResult(result)
			break
		}
	}
}

func testStopPipeline(tm *tools.ToolManager, scanner *bufio.Scanner) {
	fmt.Println(titleStyle.Render("Testing stop_pipeline"))
	fmt.Println(errorStyle.Render("This will actually stop a running pipeline!"))

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
		fmt.Println(errorStyle.Render("Repository name/ID required"))
		return
	}

	fmt.Print(infoStyle.Render("Enter pipeline number: "))
	if !scanner.Scan() {
		return
	}

	pipelineInput := strings.TrimSpace(scanner.Text())
	pipelineNum, err := strconv.Atoi(pipelineInput)
	if err != nil {
		fmt.Println(errorStyle.Render("Invalid pipeline number"))
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
				fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err)))
				return
			}

			displayResult(result)
			break
		}
	}
}

func testApprovePipeline(tm *tools.ToolManager, scanner *bufio.Scanner) {
	fmt.Println(titleStyle.Render("Testing approve_pipeline"))
	fmt.Println(errorStyle.Render("This will actually approve a pending pipeline!"))

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
		fmt.Println(errorStyle.Render("Repository name/ID required"))
		return
	}

	fmt.Print(infoStyle.Render("Enter pipeline number: "))
	if !scanner.Scan() {
		return
	}

	pipelineInput := strings.TrimSpace(scanner.Text())
	pipelineNum, err := strconv.Atoi(pipelineInput)
	if err != nil {
		fmt.Println(errorStyle.Render("Invalid pipeline number"))
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
				fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err)))
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
		fmt.Println(errorStyle.Render("Repository name/ID required"))
		return
	}

	fmt.Print(infoStyle.Render("Enter pipeline number: "))
	if !scanner.Scan() {
		return
	}

	pipelineInput := strings.TrimSpace(scanner.Text())
	pipelineNum, err := strconv.Atoi(pipelineInput)
	if err != nil {
		fmt.Println(errorStyle.Render("Invalid pipeline number"))
		return
	}

	fmt.Print(infoStyle.Render("Enter step ID: "))
	if !scanner.Scan() {
		return
	}

	stepInput := strings.TrimSpace(scanner.Text())
	stepID, err := strconv.Atoi(stepInput)
	if err != nil {
		fmt.Println(errorStyle.Render("Invalid step ID"))
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
				fmt.Println(errorStyle.Render(fmt.Sprintf("Error: %v", err)))
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

	fmt.Println(infoStyle.Render("Basic tests completed. Interactive tests require specific repository information."))
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
		fmt.Println(errorStyle.Render("Tool returned error:"))
	} else {
		fmt.Println(successStyle.Render("Tool result:"))
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
