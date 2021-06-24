package dataart

import (
	"errors"
	"net/http"
	"time"
)

// ClientConfig is the container for all required settings for DataArt Go client.
type ClientConfig struct {
	// baseURL is the server endpoint root address.
	baseURL string

	// APIKey is the authorization key for sending requests. You can find this value
	// in your dashboard. Contact support if you need help.
	APIKey string

	// FlushBufferSize is the number of pending requests held in memory. If you hit this
	// many requests, future ones will be blocking. Modify this accordingly with your workload.
	FlushBufferSize int

	// FlushNumWorkers is the total number of workers sending async requests. Each worker
	// will create a goroutine with a small footprint but beware of large values.
	// Modify this accordingly with your workload.
	FlushNumWorkers int

	// FlushNumRetries is the number of times each request is tried before giving up.
	FlushNumRetries int

	// FlushBackoffRatio is the constant time added for each execution retry. For instance
	// a FlushNumRetries of 3 and FlushBackoffRatio of 5 will cause in 5, 10, 15 seconds
	// of delay before giving up.
	FlushBackoffRatio int

	// FlushActionsBatchSize is the number of action events in batch request. If you emit
	// this much actions, a request will be created and sent.
	FlushActionsBatchSize int

	// FlushInterval is the timer duration for flushing actions. If this much time is passed
	// and there's some actions left, they will be sent to server.
	FlushInterval time.Duration

	// HTTPClient is used for executing HTTP requests. You can provide http.DefaultClient if
	// is suffices your needs.
	HTTPClient *http.Client
}

func validateConfig(cfg ClientConfig) error {
	if len(cfg.APIKey) == 0 {
		return errors.New("APIKey must not be empty")
	}

	if cfg.FlushBufferSize < 1 {
		return errors.New("FlushBufferSize can't be less than 1")
	}

	if cfg.FlushNumWorkers < 1 {
		return errors.New("FlushNumWorkers can't be less than 1")
	}

	if cfg.FlushNumRetries < 0 {
		return errors.New("FlushNumRetries must be at least 0")
	}

	if cfg.FlushBackoffRatio < 1 {
		return errors.New("FlushBackoffRatio can't be less than 1")
	}

	if cfg.FlushActionsBatchSize < 1 {
		return errors.New("FlushActionsBatchSize can't be less than 1")
	}

	if cfg.FlushInterval < time.Duration(time.Second*5) {
		return errors.New("FlushInterval must be greater than 5 seconds")
	}

	if cfg.HTTPClient == nil {
		return errors.New("HTTPClient can't be nil")
	}

	return nil
}
