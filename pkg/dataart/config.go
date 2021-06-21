package dataart

import (
	"net/http"
	"time"
)

type ClientConfig struct {
	APIKey        string
	FlushCap      int
	FlushInterval time.Duration
	HTTPClient    *http.Client
}
