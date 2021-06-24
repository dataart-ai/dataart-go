package dataart

import (
	"encoding/json"
	gohttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dataart-ai/dataart-go/internal/http"
)

type testAcceptingActionsHandler struct {
	errCh chan error
}

func (t *testAcceptingActionsHandler) ServeHTTP(w gohttp.ResponseWriter, r *gohttp.Request) {
	a := http.ActionsContainer{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&a)
	if err != nil {
		t.errCh <- err
		return
	}

	w.WriteHeader(gohttp.StatusOK)
	w.Write(nil)
}

type testAcceptingIdentitiesHandler struct {
	errCh chan error
}

func (t *testAcceptingIdentitiesHandler) ServeHTTP(w gohttp.ResponseWriter, r *gohttp.Request) {
	a := http.IdentityContainer{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&a)
	if err != nil {
		t.errCh <- err
		return
	}

	w.WriteHeader(gohttp.StatusOK)
	w.Write(nil)
}

func TestClient_WithActionsRequest(t *testing.T) {
	errCh := make(chan error)
	var actionsErr error = nil

	s := httptest.NewServer(&testAcceptingActionsHandler{errCh})
	defer s.Close()

	cfg := ClientConfig{
		baseURL:               s.URL,
		APIKey:                "api-key",
		FlushBufferSize:       2,
		FlushNumWorkers:       2,
		FlushNumRetries:       1,
		FlushBackoffRatio:     1,
		FlushActionsBatchSize: 1,
		FlushInterval:         time.Duration(5 * time.Second),
		HTTPClient:            gohttp.DefaultClient,
	}

	c, err := NewClient(cfg)
	if err != nil {
		t.Errorf("creating client failed with error: %s", err.Error())
		t.Fail()
	}
	defer c.Close()

	go func() {
		actionsErr = <-errCh
	}()

	err = c.EmitAction("event-key", "user-key", false, time.Now(), nil)
	if err != nil {
		t.Errorf("emitting action failed with error: %s", err.Error())
		t.Fail()
	}

	time.Sleep(10 * time.Millisecond)
	if actionsErr != nil {
		t.Errorf("malformed payload sent to server with error: %s", actionsErr.Error())
		t.Fail()
	}
}

func TestClient_WithIdentityRequest(t *testing.T) {
	errCh := make(chan error)
	var identitiesErr error = nil

	s := httptest.NewServer(&testAcceptingIdentitiesHandler{errCh})
	defer s.Close()

	cfg := ClientConfig{
		baseURL:               s.URL,
		APIKey:                "api-key",
		FlushBufferSize:       2,
		FlushNumWorkers:       2,
		FlushNumRetries:       1,
		FlushBackoffRatio:     1,
		FlushActionsBatchSize: 1,
		FlushInterval:         time.Duration(5 * time.Second),
		HTTPClient:            gohttp.DefaultClient,
	}

	c, err := NewClient(cfg)
	if err != nil {
		t.Errorf("creating client failed with error: %s", err.Error())
		t.Fail()
	}
	defer c.Close()

	go func() {
		identitiesErr = <-errCh
	}()

	err = c.Identify("user-key", nil)
	if err != nil {
		t.Errorf("emitting action failed with error: %s", err.Error())
		t.Fail()
	}

	time.Sleep(10 * time.Millisecond)
	if identitiesErr != nil {
		t.Errorf("malformed payload sent to server with error: %s", identitiesErr.Error())
		t.Fail()
	}
}
