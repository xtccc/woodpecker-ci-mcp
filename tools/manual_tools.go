package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sirupsen/logrus"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"

	"github.com/denysvitali/woodpecker-ci-mcp/internal/client"
)

type ToolManager struct {
	client *client.Client
	logger *logrus.Logger
	tools  []mcp.Tool
}

func NewToolManager(wclient *client.Client, logger *logrus.Logger) *ToolManager {
	tm := &ToolManager{
		client: wclient,
		logger: logger,
	}

	tm.initializeTools()
	return tm
}

func (tm *ToolManager) initializeTools() {
	tm.tools = []mcp.Tool{
		{
			Name:        "list_repositories",
			Description: "List all repositories accessible to the authenticated user",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"all": map[string]interface{}{
						"type":        "boolean",
						"description": "Include all repositories (default: false, only active repositories)",
					},
				},
			},
		},
		{
			Name:        "get_repository",
			Description: "Get detailed information about a specific repository",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"repo_id": map[string]interface{}{
						"type":        "number",
						"description": "Repository ID (use either repo_id or repo_name)",
					},
					"repo_name": map[string]interface{}{
						"type":        "string",
						"description": "Repository full name (owner/repo, use either repo_id or repo_name)",
					},
				},
			},
		},
		{
			Name:        "list_pipelines",
			Description: "List pipelines for a repository",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"repo_id": map[string]interface{}{
						"type":        "number",
						"description": "Repository ID (use either repo_id or repo_name)",
					},
					"repo_name": map[string]interface{}{
						"type":        "string",
						"description": "Repository full name (owner/repo, use either repo_id or repo_name)",
					},
				},
			},
		},
		{
			Name:        "get_pipeline_status",
			Description: "Get the status of a specific pipeline",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"repo_id": map[string]interface{}{
						"type":        "number",
						"description": "Repository ID (use either repo_id or repo_name)",
					},
					"repo_name": map[string]interface{}{
						"type":        "string",
						"description": "Repository full name (owner/repo, use either repo_id or repo_name)",
					},
					"pipeline_number": map[string]interface{}{
						"type":        "number",
						"description": "Pipeline number (required if not using 'latest')",
					},
					"latest": map[string]interface{}{
						"type":        "boolean",
						"description": "Get the latest pipeline (default: false)",
					},
				},
			},
		},
		{
			Name:        "start_pipeline",
			Description: "Start (restart) a specific pipeline",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"repo_id": map[string]interface{}{
						"type":        "number",
						"description": "Repository ID (use either repo_id or repo_name)",
					},
					"repo_name": map[string]interface{}{
						"type":        "string",
						"description": "Repository full name (owner/repo, use either repo_id or repo_name)",
					},
					"pipeline_number": map[string]interface{}{
						"type":        "number",
						"description": "Pipeline number to restart",
					},
					"fork": map[string]interface{}{
						"type":        "boolean",
						"description": "Fork the pipeline (default: false)",
					},
				},
			},
		},
		{
			Name:        "stop_pipeline",
			Description: "Stop a running pipeline",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"repo_id": map[string]interface{}{
						"type":        "number",
						"description": "Repository ID (use either repo_id or repo_name)",
					},
					"repo_name": map[string]interface{}{
						"type":        "string",
						"description": "Repository full name (owner/repo, use either repo_id or repo_name)",
					},
					"pipeline_number": map[string]interface{}{
						"type":        "number",
						"description": "Pipeline number to stop",
					},
				},
			},
		},
		{
			Name:        "approve_pipeline",
			Description: "Approve a pending pipeline",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"repo_id": map[string]interface{}{
						"type":        "number",
						"description": "Repository ID (use either repo_id or repo_name)",
					},
					"repo_name": map[string]interface{}{
						"type":        "string",
						"description": "Repository full name (owner/repo, use either repo_id or repo_name)",
					},
					"pipeline_number": map[string]interface{}{
						"type":        "number",
						"description": "Pipeline number to approve",
					},
				},
			},
		},
		{
			Name:        "get_logs",
			Description: "Get logs for a specific pipeline step",
			InputSchema: mcp.ToolInputSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"repo_id": map[string]interface{}{
						"type":        "number",
						"description": "Repository ID (use either repo_id or repo_name)",
					},
					"repo_name": map[string]interface{}{
						"type":        "string",
						"description": "Repository full name (owner/repo, use either repo_id or repo_name)",
					},
					"pipeline_number": map[string]interface{}{
						"type":        "number",
						"description": "Pipeline number",
					},
					"step_id": map[string]interface{}{
						"type":        "number",
						"description": "Step ID to get logs for",
					},
					"format": map[string]interface{}{
						"type":        "string",
						"description": "Output format: 'json' for structured data or 'text' for plain text (default: json)",
					},
				},
			},
		},
	}
}

func (tm *ToolManager) GetServerTools() []server.ServerTool {
	var serverTools []server.ServerTool
	for _, tool := range tm.tools {
		serverTools = append(serverTools, server.ServerTool{
			Tool:    tool,
			Handler: tm.getToolHandler(tool.Name),
		})
	}
	return serverTools
}

func (tm *ToolManager) getToolHandler(name string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		tm.logger.WithFields(logrus.Fields{
			"tool":      name,
			"arguments": request.Params.Arguments,
		}).Debug("Calling tool")

		// Type assert arguments to map[string]interface{}
		arguments, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: "Invalid arguments format",
					},
				},
				IsError: true,
			}, nil
		}

		switch name {
		case "list_repositories":
			return tm.handleListRepositories(ctx, arguments)
		case "get_repository":
			return tm.handleGetRepository(ctx, arguments)
		case "list_pipelines":
			return tm.handleListPipelines(ctx, arguments)
		case "get_pipeline_status":
			return tm.handleGetPipelineStatus(ctx, arguments)
		case "start_pipeline":
			return tm.handleStartPipeline(ctx, arguments)
		case "stop_pipeline":
			return tm.handleStopPipeline(ctx, arguments)
		case "approve_pipeline":
			return tm.handleApprovePipeline(ctx, arguments)
		case "get_logs":
			return tm.handleGetLogs(ctx, arguments)
		default:
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf("Unknown tool: %s", name),
					},
				},
				IsError: true,
			}, nil
		}
	}
}

func (tm *ToolManager) handleListRepositories(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	repositories, err := tm.client.ListRepositories()
	if err != nil {
		tm.logger.WithError(err).Error("Failed to list repositories")
		return tm.errorResult(fmt.Sprintf("Failed to list repositories: %v", err)), nil
	}

	showAll := getBool(arguments, "all", false)
	if !showAll {
		var activeRepos []*woodpecker.Repo
		for _, repo := range repositories {
			if repo.IsActive {
				activeRepos = append(activeRepos, repo)
			}
		}
		repositories = activeRepos
	}

	response := map[string]interface{}{
		"repositories": repositories,
		"total_count":  len(repositories),
		"showing_all":  showAll,
	}

	return tm.jsonResult(response)
}

func (tm *ToolManager) handleGetRepository(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	repoID, err := getRepoID(tm.client, arguments)
	if err != nil {
		return tm.errorResult(err.Error()), nil
	}

	repo, err := tm.client.GetRepository(repoID)
	if err != nil {
		return tm.errorResult(fmt.Sprintf("Failed to get repository: %v", err)), nil
	}

	return tm.jsonResult(repo)
}

func (tm *ToolManager) handleListPipelines(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	repoID, err := getRepoID(tm.client, arguments)
	if err != nil {
		return tm.errorResult(err.Error()), nil
	}

	pipelines, err := tm.client.ListPipelines(repoID)
	if err != nil {
		return tm.errorResult(fmt.Sprintf("Failed to list pipelines: %v", err)), nil
	}

	response := map[string]interface{}{
		"repo_id":     repoID,
		"pipelines":   pipelines,
		"total_count": len(pipelines),
	}

	return tm.jsonResult(response)
}

func (tm *ToolManager) handleGetPipelineStatus(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	repoID, err := getRepoID(tm.client, arguments)
	if err != nil {
		return tm.errorResult(err.Error()), nil
	}

	var pipeline interface{}
	latest := getBool(arguments, "latest", false)

	if latest {
		pipeline, err = tm.client.GetLastPipeline(repoID)
	} else {
		pipelineNum, err := requireNumber(arguments, "pipeline_number")
		if err != nil {
			return tm.errorResult("Either pipeline_number or latest=true must be provided"), nil
		}
		pipeline, err = tm.client.GetPipeline(repoID, int64(pipelineNum))
	}

	if err != nil {
		return tm.errorResult(fmt.Sprintf("Failed to get pipeline: %v", err)), nil
	}

	return tm.jsonResult(pipeline)
}

func (tm *ToolManager) handleStartPipeline(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	repoID, err := getRepoID(tm.client, arguments)
	if err != nil {
		return tm.errorResult(err.Error()), nil
	}

	pipelineNum, err := requireNumber(arguments, "pipeline_number")
	if err != nil {
		return tm.errorResult(err.Error()), nil
	}

	params := make(map[string]string)
	if getBool(arguments, "fork", false) {
		params["fork"] = "true"
	}

	pipeline, err := tm.client.StartPipeline(repoID, int64(pipelineNum), params)
	if err != nil {
		return tm.errorResult(fmt.Sprintf("Failed to start pipeline: %v", err)), nil
	}

	return tm.jsonResult(pipeline)
}

func (tm *ToolManager) handleStopPipeline(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	repoID, err := getRepoID(tm.client, arguments)
	if err != nil {
		return tm.errorResult(err.Error()), nil
	}

	pipelineNum, err := requireNumber(arguments, "pipeline_number")
	if err != nil {
		return tm.errorResult(err.Error()), nil
	}

	err = tm.client.StopPipeline(repoID, int64(pipelineNum))
	if err != nil {
		return tm.errorResult(fmt.Sprintf("Failed to stop pipeline: %v", err)), nil
	}

	response := map[string]interface{}{
		"success":         true,
		"message":         "Pipeline stopped successfully",
		"repo_id":         repoID,
		"pipeline_number": int64(pipelineNum),
	}

	return tm.jsonResult(response)
}

func (tm *ToolManager) handleApprovePipeline(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	repoID, err := getRepoID(tm.client, arguments)
	if err != nil {
		return tm.errorResult(err.Error()), nil
	}

	pipelineNum, err := requireNumber(arguments, "pipeline_number")
	if err != nil {
		return tm.errorResult(err.Error()), nil
	}

	pipeline, err := tm.client.ApprovePipeline(repoID, int64(pipelineNum))
	if err != nil {
		return tm.errorResult(fmt.Sprintf("Failed to approve pipeline: %v", err)), nil
	}

	return tm.jsonResult(pipeline)
}

func (tm *ToolManager) handleGetLogs(ctx context.Context, arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	repoID, err := getRepoID(tm.client, arguments)
	if err != nil {
		return tm.errorResult(err.Error()), nil
	}

	pipelineNum, err := requireNumber(arguments, "pipeline_number")
	if err != nil {
		return tm.errorResult(err.Error()), nil
	}

	stepID, err := requireNumber(arguments, "step_id")
	if err != nil {
		return tm.errorResult(err.Error()), nil
	}

	format := getString(arguments, "format", "json")

	logs, err := tm.client.GetStepLogs(repoID, int64(pipelineNum), int64(stepID))
	if err != nil {
		return tm.errorResult(fmt.Sprintf("Failed to get logs: %v", err)), nil
	}

	if format == "text" {
		var logLines []string
		for _, logEntry := range logs {
			if len(logEntry.Data) > 0 {
				logLines = append(logLines, string(logEntry.Data))
			}
		}

		plainText := fmt.Sprintf("Logs for repo %d, pipeline %d, step %d:\n%s",
			repoID, int64(pipelineNum), int64(stepID), strings.Join(logLines, "\n"))

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: plainText,
				},
			},
		}, nil
	}

	response := map[string]interface{}{
		"repo_id":         repoID,
		"pipeline_number": int64(pipelineNum),
		"step_id":         int64(stepID),
		"logs":            logs,
		"log_count":       len(logs),
	}

	return tm.jsonResult(response)
}

func (tm *ToolManager) jsonResult(data interface{}) (*mcp.CallToolResult, error) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return tm.errorResult(fmt.Sprintf("Failed to format response: %v", err)), nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(jsonData),
			},
		},
	}, nil
}

func (tm *ToolManager) errorResult(message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: message,
			},
		},
		IsError: true,
	}
}
