package dataart

import (
	"errors"
	"time"

	"github.com/dataart-ai/dataart-go/internal/http"
)

type Client struct {
	Config ClientConfig
	Tracker
}

func (c *Client) Close() error {
	return nil
}

func DefaultClient(cfg ClientConfig) (*Client, error) {
	if len(cfg.APIKey) == 0 {
		return nil, errors.New("APIKey must not be empty")
	}

	if cfg.FlushCap < 1 {
		return nil, errors.New("FlushCap must be greater than zero")
	}

	if cfg.FlushInterval < time.Duration(time.Second*5) {
		return nil, errors.New("FlushInterval must be greater than 5 seconds")
	}

	trk := &trackerImpl{
		uploader: http.NewUploader(cfg.APIKey, cfg.HTTPClient),
	}

	c := &Client{
		Config:  cfg,
		Tracker: trk,
	}

	return c, nil
}
