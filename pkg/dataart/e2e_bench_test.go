package dataart

import (
	gohttp "net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type testAcceptingBenchHandler struct{}

func (t *testAcceptingBenchHandler) ServeHTTP(w gohttp.ResponseWriter, r *gohttp.Request) {
	w.WriteHeader(gohttp.StatusOK)
	w.Write(nil)
}

func BenchmarkClient(b *testing.B) {
	s := httptest.NewServer(&testAcceptingBenchHandler{})
	defer s.Close()

	cfg := ClientConfig{
		baseURL:               s.URL,
		APIKey:                "some-api-key",
		FlushBufferSize:       100,
		FlushNumWorkers:       16,
		FlushNumRetries:       3,
		FlushBackoffRatio:     5,
		FlushActionsBatchSize: 20,
		FlushInterval:         time.Duration(5 * time.Second),
		HTTPClient:            gohttp.DefaultClient,
	}

	c, err := NewClient(cfg)
	if err != nil {
		b.Errorf("creating client failed with error: %s", err.Error())
		b.Fail()
	}
	defer c.Close()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			err = c.EmitAction("event-key", "user-key", false, time.Now(), nil)
			if err != nil {
				b.Errorf("emitting action failed with error: %s", err.Error())
				b.Fail()
			}
		}
	})

}
