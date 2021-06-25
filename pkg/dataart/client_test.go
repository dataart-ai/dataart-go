package dataart

import (
	gohttp "net/http"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	var err error
	var cfg ClientConfig

	cfg = ClientConfig{
		APIKey:                "",
		FlushBufferSize:       2,
		FlushNumWorkers:       2,
		FlushNumRetries:       1,
		FlushBackoffRatio:     1,
		FlushActionsBatchSize: 1,
		FlushInterval:         time.Duration(5 * time.Second),
		HTTPClient:            gohttp.DefaultClient,
	}
	// APIKey
	_, err = NewClient(cfg)
	if err == nil {
		t.Fail()
	}

	cfg = ClientConfig{
		APIKey:                "api-key",
		FlushBufferSize:       0,
		FlushNumWorkers:       2,
		FlushNumRetries:       1,
		FlushBackoffRatio:     1,
		FlushActionsBatchSize: 1,
		FlushInterval:         time.Duration(5 * time.Second),
		HTTPClient:            gohttp.DefaultClient,
	}
	// FlushBufferSize
	_, err = NewClient(cfg)
	if err == nil {
		t.Fail()
	}

	cfg = ClientConfig{
		APIKey:                "api-key",
		FlushBufferSize:       1,
		FlushNumWorkers:       0,
		FlushNumRetries:       1,
		FlushBackoffRatio:     1,
		FlushActionsBatchSize: 1,
		FlushInterval:         time.Duration(5 * time.Second),
		HTTPClient:            gohttp.DefaultClient,
	}
	// FlushNumWorkers
	_, err = NewClient(cfg)
	if err == nil {
		t.Fail()
	}

	cfg = ClientConfig{
		APIKey:                "api-key",
		FlushBufferSize:       1,
		FlushNumWorkers:       1,
		FlushNumRetries:       -1,
		FlushBackoffRatio:     1,
		FlushActionsBatchSize: 1,
		FlushInterval:         time.Duration(5 * time.Second),
		HTTPClient:            gohttp.DefaultClient,
	}
	// FlushNumRetries
	_, err = NewClient(cfg)
	if err == nil {
		t.Fail()
	}

	cfg = ClientConfig{
		APIKey:                "api-key",
		FlushBufferSize:       1,
		FlushNumWorkers:       1,
		FlushNumRetries:       1,
		FlushBackoffRatio:     0,
		FlushActionsBatchSize: 1,
		FlushInterval:         time.Duration(5 * time.Second),
		HTTPClient:            gohttp.DefaultClient,
	}
	// FlushBackoffRatio
	_, err = NewClient(cfg)
	if err == nil {
		t.Fail()
	}

	cfg = ClientConfig{
		APIKey:                "api-key",
		FlushBufferSize:       1,
		FlushNumWorkers:       1,
		FlushNumRetries:       1,
		FlushBackoffRatio:     1,
		FlushActionsBatchSize: 0,
		FlushInterval:         time.Duration(5 * time.Second),
		HTTPClient:            gohttp.DefaultClient,
	}
	// FlushActionsBatchSize
	_, err = NewClient(cfg)
	if err == nil {
		t.Fail()
	}

	cfg = ClientConfig{
		APIKey:                "api-key",
		FlushBufferSize:       1,
		FlushNumWorkers:       1,
		FlushNumRetries:       1,
		FlushBackoffRatio:     1,
		FlushActionsBatchSize: 1,
		FlushInterval:         time.Duration(4 * time.Second),
		HTTPClient:            gohttp.DefaultClient,
	}
	// FlushInterval
	_, err = NewClient(cfg)
	if err == nil {
		t.Fail()
	}

	cfg = ClientConfig{
		APIKey:                "api-key",
		FlushBufferSize:       1,
		FlushNumWorkers:       1,
		FlushNumRetries:       1,
		FlushBackoffRatio:     1,
		FlushActionsBatchSize: 1,
		FlushInterval:         time.Duration(5 * time.Second),
		HTTPClient:            nil,
	}
	// HTTPClient
	_, err = NewClient(cfg)
	if err == nil {
		t.Fail()
	}
}

func TestClient_WithEmitActionAndInvalidData(t *testing.T) {
	t.Parallel()

	cfg := ClientConfig{
		APIKey:                "api-key",
		FlushBufferSize:       1,
		FlushNumWorkers:       1,
		FlushNumRetries:       1,
		FlushBackoffRatio:     1,
		FlushActionsBatchSize: 1,
		FlushInterval:         time.Duration(5 * time.Second),
		HTTPClient:            gohttp.DefaultClient,
	}

	c, _ := NewClient(cfg)
	err := c.EmitAction("", "", true, time.Now(), nil)
	if err == nil {
		t.Fail()
	}
}

func TestClient_WithIdentifyAndInvalidData(t *testing.T) {
	t.Parallel()

	cfg := ClientConfig{
		APIKey:                "api-key",
		FlushBufferSize:       1,
		FlushNumWorkers:       1,
		FlushNumRetries:       1,
		FlushBackoffRatio:     1,
		FlushActionsBatchSize: 1,
		FlushInterval:         time.Duration(5 * time.Second),
		HTTPClient:            gohttp.DefaultClient,
	}

	c, _ := NewClient(cfg)
	err := c.Identify("", nil)
	if err == nil {
		t.Fail()
	}
}
