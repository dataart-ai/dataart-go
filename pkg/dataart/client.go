package dataart

import (
	"errors"
	"time"

	"github.com/dataart-ai/dataart-go/internal/http"
	"github.com/dataart-ai/dataart-go/internal/task"
)

const (
	sourcingURL = "https://sourcing.datartproject.com"
)

type httpUploader interface {
	UploadAction(cnt http.ActionContainer) error
	UploadIdentity(cnt http.IdentityContainer) error
	Shutdown()
}

// Client encapsulates a DataArt client.
type Client struct {
	Config ClientConfig
	hu     httpUploader
}

// EmitAction creates an action object with given properties and uploads it to server.
func (c *Client) EmitAction(key string, userKey string, isAnonymousUser bool,
	timestamp time.Time, metadata map[string]interface{}) error {

	if len(key) == 0 {
		return errors.New("event key identifier must not empty")
	}

	return c.hu.UploadAction(
		http.ActionContainer{
			Key:             key,
			UserKey:         userKey,
			IsAnonymousUser: isAnonymousUser,
			Timestamp:       timestamp,
			Metadata:        metadata,
		},
	)
}

// Identify creates an identity object with given properties and uploads it to server.
func (c *Client) Identify(userKey string, metadata map[string]interface{}) error {
	if len(userKey) == 0 {
		return errors.New("userKey must not empty")
	}

	return c.hu.UploadIdentity(
		http.IdentityContainer{
			UserKey:  userKey,
			Metadata: metadata,
		},
	)
}

// Close gracefully terminates the underlying dependencies.
func (c *Client) Close() {
	c.hu.Shutdown()
}

// NewClient creates a new Client instance with given configuration values. Use this
// function to instantiate a concrete Client type.
func NewClient(cfg ClientConfig) (*Client, error) {
	err := validateConfig(cfg)
	if err != nil {
		return nil, err
	}

	if len(cfg.baseURL) == 0 {
		cfg.baseURL = sourcingURL
	}

	tm, err := task.NewManager(cfg.FlushNumWorkers, cfg.FlushBufferSize,
		cfg.FlushNumRetries, cfg.FlushBackoffRatio, nil, nil)

	if err != nil {
		return nil, err
	}

	uploader, err := http.NewUploader(cfg.baseURL, cfg.APIKey,
		cfg.FlushActionsBatchSize, cfg.FlushInterval, cfg.HTTPClient, tm)

	if err != nil {
		return nil, err
	}

	c := &Client{
		Config: cfg,
		hu:     uploader,
	}

	return c, nil
}
