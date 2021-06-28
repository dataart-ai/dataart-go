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
	_, err = NewClient(cfg)
	if err == nil {
		t.Error("given APIKey is invalid")
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
	_, err = NewClient(cfg)
	if err == nil {
		t.Error("given FlushBufferSize is invalid")
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
	_, err = NewClient(cfg)
	if err == nil {
		t.Error("given FlushNumWorkers is invalid")
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
	_, err = NewClient(cfg)
	if err == nil {
		t.Error("given FlushNumRetries is invalid")
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
	_, err = NewClient(cfg)
	if err == nil {
		t.Error("given FlushBackoffRatio is invalid")
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
	_, err = NewClient(cfg)
	if err == nil {
		t.Error("given FlushActionsBatchSize is invalid")
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
	_, err = NewClient(cfg)
	if err == nil {
		t.Error("given FlushInterval is invalid")
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
	_, err = NewClient(cfg)
	if err == nil {
		t.Error("given HTTPClient is invalid")
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
		t.Error("given eventKey is invalid")
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
		t.Error("given user key is invalid")
		t.Fail()
	}
}
