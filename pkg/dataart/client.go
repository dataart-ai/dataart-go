package dataart

import (
	"github.com/dataart-ai/dataart-go/internal/http"
	"github.com/dataart-ai/dataart-go/internal/task"
)

// Client is a DataArt HTTP API client.
type Client struct {
	Config ClientConfig
	Tracker
}

// Close gracefully terminates the underlying tracker instance.
func (c *Client) Close() {
	c.Tracker.Close()
}

// NewClient initiates the a Client with given configuration values.
func NewClient(cfg ClientConfig) (*Client, error) {
	err := validateConfig(cfg)
	if err != nil {
		return nil, err
	}

	if len(cfg.baseURL) == 0 {
		cfg.baseURL = "http://sourcing.datartproject.com"
	}

	tm, err := task.NewManager(cfg.FlushNumWorkers, cfg.FlushBufferSize,
		cfg.FlushNumRetries, cfg.FlushBackoffRatio, nil, nil)

	if err != nil {
		return nil, err
	}

	up, err := http.NewUploader(cfg.baseURL, cfg.APIKey,
		cfg.FlushActionsBatchSize, cfg.FlushInterval, cfg.HTTPClient, tm)

	if err != nil {
		return nil, err
	}

	trk := &trackerImpl{
		uploader: up,
	}

	c := &Client{
		Config:  cfg,
		Tracker: trk,
	}

	return c, nil
}
