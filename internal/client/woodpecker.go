package client

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
	"golang.org/x/oauth2"
)

type Client struct {
	client woodpecker.Client
	logger *logrus.Logger
	url    string
}

type Config struct {
	URL   string
	Token string
}

func New(cfg Config, logger *logrus.Logger) (*Client, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("woodpecker URL is required")
	}
	if cfg.Token == "" {
		return nil, fmt.Errorf("woodpecker token is required")
	}

	// Create OAuth2 config and client
	config := &oauth2.Config{}
	token := &oauth2.Token{AccessToken: cfg.Token}

	ctx := context.Background()
	httpClient := config.Client(ctx, token)

	// Set timeout for HTTP client
	httpClient.Timeout = 30 * time.Second

	// Create Woodpecker client
	client := woodpecker.NewClient(cfg.URL, httpClient)

	wclient := &Client{
		client: client,
		logger: logger,
		url:    cfg.URL,
	}

	// Test connection
	if err := wclient.TestConnection(); err != nil {
		return nil, fmt.Errorf("failed to connect to Woodpecker server: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"url": cfg.URL,
	}).Info("Successfully connected to Woodpecker CI server")

	return wclient, nil
}

func (c *Client) TestConnection() error {
	// Try to get current user info to test connection
	_, err := c.client.Self()
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	return nil
}

func (c *Client) ListRepositories() ([]*woodpecker.Repo, error) {
	repos, err := c.client.RepoList(woodpecker.RepoListOptions{})
	if err != nil {
		c.logger.WithError(err).Error("Failed to list repositories")
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}

	c.logger.WithField("count", len(repos)).Debug("Listed repositories")
	return repos, nil
}

func (c *Client) GetRepository(repoID int64) (*woodpecker.Repo, error) {
	repo, err := c.client.Repo(repoID)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"repo_id": repoID,
			"error":   err,
		}).Error("Failed to get repository")
		return nil, fmt.Errorf("failed to get repository %d: %w", repoID, err)
	}

	return repo, nil
}

func (c *Client) LookupRepository(fullName string) (*woodpecker.Repo, error) {
	repo, err := c.client.RepoLookup(fullName)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"repo_name": fullName,
			"error":     err,
		}).Error("Failed to lookup repository")
		return nil, fmt.Errorf("failed to lookup repository %s: %w", fullName, err)
	}

	return repo, nil
}

func (c *Client) ListPipelines(repoID int64) ([]*woodpecker.Pipeline, error) {
	pipelines, err := c.client.PipelineList(repoID, woodpecker.PipelineListOptions{})
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"repo_id": repoID,
			"error":   err,
		}).Error("Failed to list pipelines")
		return nil, fmt.Errorf("failed to list pipelines for repo %d: %w", repoID, err)
	}

	c.logger.WithFields(logrus.Fields{
		"repo_id": repoID,
		"count":   len(pipelines),
	}).Debug("Listed pipelines")
	return pipelines, nil
}

func (c *Client) GetPipeline(repoID, pipelineNum int64) (*woodpecker.Pipeline, error) {
	pipeline, err := c.client.Pipeline(repoID, pipelineNum)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"repo_id":      repoID,
			"pipeline_num": pipelineNum,
			"error":        err,
		}).Error("Failed to get pipeline")
		return nil, fmt.Errorf("failed to get pipeline %d for repo %d: %w", pipelineNum, repoID, err)
	}

	return pipeline, nil
}

func (c *Client) GetLastPipeline(repoID int64) (*woodpecker.Pipeline, error) {
	pipeline, err := c.client.PipelineLast(repoID, woodpecker.PipelineLastOptions{})
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"repo_id": repoID,
			"error":   err,
		}).Error("Failed to get last pipeline")
		return nil, fmt.Errorf("failed to get last pipeline for repo %d: %w", repoID, err)
	}

	return pipeline, nil
}

func (c *Client) StartPipeline(repoID, pipelineNum int64, params map[string]string) (*woodpecker.Pipeline, error) {
	options := woodpecker.PipelineStartOptions{
		Params: params,
	}
	pipeline, err := c.client.PipelineStart(repoID, pipelineNum, options)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"repo_id":      repoID,
			"pipeline_num": pipelineNum,
			"error":        err,
		}).Error("Failed to start pipeline")
		return nil, fmt.Errorf("failed to start pipeline %d for repo %d: %w", pipelineNum, repoID, err)
	}

	c.logger.WithFields(logrus.Fields{
		"repo_id":      repoID,
		"pipeline_num": pipelineNum,
	}).Info("Started pipeline")
	return pipeline, nil
}

func (c *Client) StopPipeline(repoID, pipelineNum int64) error {
	err := c.client.PipelineStop(repoID, pipelineNum)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"repo_id":      repoID,
			"pipeline_num": pipelineNum,
			"error":        err,
		}).Error("Failed to stop pipeline")
		return fmt.Errorf("failed to stop pipeline %d for repo %d: %w", pipelineNum, repoID, err)
	}

	c.logger.WithFields(logrus.Fields{
		"repo_id":      repoID,
		"pipeline_num": pipelineNum,
	}).Info("Stopped pipeline")
	return nil
}

func (c *Client) ApprovePipeline(repoID, pipelineNum int64) (*woodpecker.Pipeline, error) {
	pipeline, err := c.client.PipelineApprove(repoID, pipelineNum)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"repo_id":      repoID,
			"pipeline_num": pipelineNum,
			"error":        err,
		}).Error("Failed to approve pipeline")
		return nil, fmt.Errorf("failed to approve pipeline %d for repo %d: %w", pipelineNum, repoID, err)
	}

	c.logger.WithFields(logrus.Fields{
		"repo_id":      repoID,
		"pipeline_num": pipelineNum,
	}).Info("Approved pipeline")
	return pipeline, nil
}

func (c *Client) DeclinePipeline(repoID, pipelineNum int64) (*woodpecker.Pipeline, error) {
	pipeline, err := c.client.PipelineDecline(repoID, pipelineNum)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"repo_id":      repoID,
			"pipeline_num": pipelineNum,
			"error":        err,
		}).Error("Failed to decline pipeline")
		return nil, fmt.Errorf("failed to decline pipeline %d for repo %d: %w", pipelineNum, repoID, err)
	}

	c.logger.WithFields(logrus.Fields{
		"repo_id":      repoID,
		"pipeline_num": pipelineNum,
	}).Info("Declined pipeline")
	return pipeline, nil
}

func (c *Client) CreatePipeline(repoID int64, opt *woodpecker.PipelineOptions) (*woodpecker.Pipeline, error) {
	pipeline, err := c.client.PipelineCreate(repoID, opt)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"repo_id": repoID,
			"error":   err,
		}).Error("Failed to create pipeline")
		return nil, fmt.Errorf("failed to create pipeline for repo %d: %w", repoID, err)
	}

	c.logger.WithFields(logrus.Fields{
		"repo_id":      repoID,
		"pipeline_id":  pipeline.ID,
		"pipeline_num": pipeline.Number,
	}).Info("Created pipeline")
	return pipeline, nil
}

// Log methods
func (c *Client) GetStepLogs(repoID, pipelineNum, stepID int64) ([]*woodpecker.LogEntry, error) {
	logs, err := c.client.StepLogEntries(repoID, pipelineNum, stepID)
	if err != nil {
		c.logger.WithFields(logrus.Fields{
			"repo_id":      repoID,
			"pipeline_num": pipelineNum,
			"step_id":      stepID,
			"error":        err,
		}).Error("Failed to get step logs")
		return nil, fmt.Errorf("failed to get logs for step %d in pipeline %d for repo %d: %w", stepID, pipelineNum, repoID, err)
	}

	return logs, nil
}

// User methods
func (c *Client) GetCurrentUser() (*woodpecker.User, error) {
	user, err := c.client.Self()
	if err != nil {
		c.logger.WithError(err).Error("Failed to get current user")
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	return user, nil
}
