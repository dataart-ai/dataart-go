package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockAcceptingHandler struct {
	feedbackCh chan bool
}

func (m *mockAcceptingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if m.feedbackCh != nil {
		m.feedbackCh <- true
	}

	w.WriteHeader(http.StatusOK)
	w.Write(nil)
}

type mockRejectingHandler struct{}

func (m *mockRejectingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write(nil)
}

type mockWorkingTaskManager struct{}

func (m *mockWorkingTaskManager) Queue(work func() error) error {
	work()
	return nil
}

func (m *mockWorkingTaskManager) Shutdown() {}

func TestNewUploader(t *testing.T) {
	t.Parallel()

	_, err := NewUploader("", "api-key", 1, time.Duration(5*time.Second), http.DefaultClient, &mockWorkingTaskManager{})
	if err == nil {
		t.Error("given baseURL is invalid")
		t.Fail()
	}

	_, err = NewUploader("localhost:9090", "", 1, time.Duration(5*time.Second), http.DefaultClient, &mockWorkingTaskManager{})
	if err == nil {
		t.Error("given apiKey is invalid")
		t.Fail()
	}

	_, err = NewUploader("https://something.com", "api-key", 0, time.Duration(5*time.Second), http.DefaultClient, &mockWorkingTaskManager{})
	if err == nil {
		t.Error("given batchSize is invalid")
		t.Fail()
	}

	_, err = NewUploader("localhost:9090", "api-key", 1, time.Duration(1*time.Second), http.DefaultClient, &mockWorkingTaskManager{})
	if err == nil {
		t.Error("given uploadInterval is invalid")
		t.Fail()
	}

	_, err = NewUploader("localhost:9090", "api-key", 1, time.Duration(5*time.Second), nil, &mockWorkingTaskManager{})
	if err == nil {
		t.Error("given httpClient is invalid")
		t.Fail()
	}

	_, err = NewUploader("localhost:9090", "api-key", 1, time.Duration(5*time.Second), http.DefaultClient, nil)
	if err == nil {
		t.Error("given TaskManager is invalid")
		t.Fail()
	}
}

func TestUploader_WithAcceptingHandlerAndUploadActions(t *testing.T) {
	t.Parallel()

	s := httptest.NewServer(&mockAcceptingHandler{nil})
	defer s.Close()

	u, _ := NewUploader(
		s.URL,
		"some-api-key",
		1,
		time.Duration(5*time.Second),
		http.DefaultClient,
		&mockWorkingTaskManager{})

	u.UploadAction(ActionContainer{
		Key:             "some-event-key",
		UserKey:         "some-user-key",
		IsAnonymousUser: false,
		Timestamp:       time.Now(),
		Metadata:        nil,
	})

	u.Shutdown()

	if len(u.actionsBatch) != 0 {
		t.Fail()
	}
}

func TestUploader_WithRejectingHandlerAndUploadActions(t *testing.T) {
	t.Parallel()

	s := httptest.NewServer(&mockRejectingHandler{})
	defer s.Close()

	u, _ := NewUploader(
		s.URL,
		"some-api-key",
		1,
		time.Duration(5*time.Second),
		http.DefaultClient,
		&mockWorkingTaskManager{})

	u.UploadAction(ActionContainer{
		Key:             "some-event-key",
		UserKey:         "some-user-key",
		IsAnonymousUser: false,
		Timestamp:       time.Now(),
		Metadata:        nil,
	})

	u.Shutdown()

	if len(u.actionsBatch) != 0 {
		t.Fail()
	}
}

func TestUploader_WithAcceptingHandlerAndUploadIdentity(t *testing.T) {
	t.Parallel()

	s := httptest.NewServer(&mockAcceptingHandler{nil})
	defer s.Close()

	u, _ := NewUploader(
		s.URL,
		"some-api-key",
		1,
		time.Duration(5*time.Second),
		http.DefaultClient,
		&mockWorkingTaskManager{})

	u.UploadIdentity(IdentityContainer{
		UserKey: "some-user-key",
	})

	u.Shutdown()

	if len(u.actionsBatch) != 0 {
		t.Fail()
	}
}

func TestUploader_WithRejectingHandlerAndUploadIdentity(t *testing.T) {
	t.Parallel()

	s := httptest.NewServer(&mockRejectingHandler{})
	defer s.Close()

	u, _ := NewUploader(
		s.URL,
		"some-api-key",
		1,
		time.Duration(5*time.Second),
		http.DefaultClient,
		&mockWorkingTaskManager{})

	u.UploadIdentity(IdentityContainer{
		UserKey: "some-user-key",
	})

	u.Shutdown()

	if len(u.actionsBatch) != 0 {
		t.Fail()
	}
}

func TestUploader_WithTimerFeedbackChannel(t *testing.T) {
	t.Parallel()

	feedbackCh := make(chan bool)
	var reqReceivedByServer bool = false

	s := httptest.NewServer(&mockAcceptingHandler{feedbackCh})
	defer s.Close()

	u, _ := NewUploader(
		s.URL,
		"some-api-key",
		1,
		time.Duration(5*time.Second),
		http.DefaultClient,
		&mockWorkingTaskManager{})

	go func() {
		reqReceivedByServer = <-feedbackCh
	}()

	u.UploadAction(ActionContainer{
		Key:             "some-event-key",
		UserKey:         "some-user-key",
		IsAnonymousUser: false,
		Timestamp:       time.Now(),
		Metadata:        nil,
	})

	time.Sleep(1500 * time.Millisecond)
	if !reqReceivedByServer {
		t.Fail()
	}

	u.Shutdown()
}

func TestUploader_WithShutdownUploader(t *testing.T) {
	t.Parallel()

	s := httptest.NewServer(&mockAcceptingHandler{nil})
	defer s.Close()

	u, _ := NewUploader(
		s.URL,
		"some-api-key",
		1,
		time.Duration(5*time.Second),
		http.DefaultClient,
		&mockWorkingTaskManager{})

	u.UploadIdentity(IdentityContainer{
		UserKey: "some-user-key",
	})

	// Multiple shutdowns should be idempotent.
	u.Shutdown()
	u.Shutdown()

	var err error
	err = u.UploadAction(ActionContainer{
		Key:             "some-event-key",
		UserKey:         "some-user-key",
		IsAnonymousUser: false,
		Timestamp:       time.Now(),
		Metadata:        nil,
	})

	if err == nil {
		t.Fail()
	}

	err = u.UploadIdentity(IdentityContainer{
		UserKey: "some-user-key",
	})

	if err == nil {
		t.Fail()
	}
}

func TestUploader_WithPrematureShutdown(t *testing.T) {
	t.Parallel()

	feedbackCh := make(chan bool)
	var reqReceivedByServer bool = false

	s := httptest.NewServer(&mockAcceptingHandler{feedbackCh})
	defer s.Close()

	u, _ := NewUploader(
		s.URL,
		"some-api-key",
		100,
		time.Duration(20*time.Second),
		http.DefaultClient,
		&mockWorkingTaskManager{})

	go func() {
		reqReceivedByServer = <-feedbackCh
	}()

	u.UploadAction(ActionContainer{
		Key:             "some-event-key",
		UserKey:         "some-user-key",
		IsAnonymousUser: false,
		Timestamp:       time.Now(),
		Metadata:        nil,
	})
	u.Shutdown()
	// Since we closed uploader immediately and there's an action remaining,
	// it should be sent to server.

	if !reqReceivedByServer {
		t.Fail()
	}
}
