package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dataart-ai/dataart-go/internal/task"
)

type testAcceptingHandler struct {
	feedbackCh chan bool
}

func (t *testAcceptingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if t.feedbackCh != nil {
		t.feedbackCh <- true
	}

	w.WriteHeader(http.StatusOK)
	w.Write(nil)
}

type testRejectingHandler struct{}

func (t *testRejectingHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write(nil)
}

type testWorkingTaskManager struct{}

func (tm *testWorkingTaskManager) Queue(t task.Task) error {
	t.Work()
	return nil
}

func (tm *testWorkingTaskManager) Shutdown() {}

func TestNewUploader(t *testing.T) {
	t.Parallel()

	_, err := NewUploader("", "api-key", 1, time.Duration(1*time.Second), http.DefaultClient, &testWorkingTaskManager{})
	if err == nil {
		t.Fail()
	}

	_, err = NewUploader("localhost:9090", "", 1, time.Duration(1*time.Second), http.DefaultClient, &testWorkingTaskManager{})
	if err == nil {
		t.Fail()
	}

	_, err = NewUploader("localhost:9090", "api-key", 0, time.Duration(1*time.Second), http.DefaultClient, &testWorkingTaskManager{})
	if err == nil {
		t.Fail()
	}

	_, err = NewUploader("localhost:9090", "api-key", 1, time.Duration(500*time.Millisecond), http.DefaultClient, &testWorkingTaskManager{})
	if err == nil {
		t.Fail()
	}

	_, err = NewUploader("localhost:9090", "api-key", 1, time.Duration(1*time.Second), nil, &testWorkingTaskManager{})
	if err == nil {
		t.Fail()
	}

	_, err = NewUploader("localhost:9090", "api-key", 1, time.Duration(1*time.Second), http.DefaultClient, nil)
	if err == nil {
		t.Fail()
	}

	_, err = NewUploader("localhost:9090", "api-key", 1, time.Duration(1*time.Second), http.DefaultClient, &testWorkingTaskManager{})
	if err != nil {
		t.Fail()
	}
}

func TestUploader_WithAcceptingHandlerAndUploadActions(t *testing.T) {
	t.Parallel()

	s := httptest.NewServer(&testAcceptingHandler{nil})
	defer s.Close()

	u := &uploaderImpl{
		baseURL:        s.URL,
		apiKey:         "some-api-key",
		batchSize:      1,
		uploadInterval: time.Duration(1 * time.Second),
		httpClient:     http.DefaultClient,
		taskManager:    &testWorkingTaskManager{},
		actionsBatch:   make([]ActionContainer, 0),
		tasks:          make(chan uploadTask),
		doneCh:         make(chan struct{}),
	}

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

	s := httptest.NewServer(&testRejectingHandler{})
	defer s.Close()

	u := &uploaderImpl{
		baseURL:        s.URL,
		apiKey:         "some-api-key",
		batchSize:      1,
		uploadInterval: time.Duration(1 * time.Second),
		httpClient:     http.DefaultClient,
		taskManager:    &testWorkingTaskManager{},
		actionsBatch:   make([]ActionContainer, 0),
		tasks:          make(chan uploadTask),
		doneCh:         make(chan struct{}),
	}

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

	s := httptest.NewServer(&testAcceptingHandler{nil})
	defer s.Close()

	u := &uploaderImpl{
		baseURL:        s.URL,
		apiKey:         "some-api-key",
		batchSize:      1,
		uploadInterval: time.Duration(1 * time.Second),
		httpClient:     http.DefaultClient,
		taskManager:    &testWorkingTaskManager{},
		actionsBatch:   make([]ActionContainer, 0),
		tasks:          make(chan uploadTask),
		doneCh:         make(chan struct{}),
	}

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

	s := httptest.NewServer(&testRejectingHandler{})
	defer s.Close()

	u := &uploaderImpl{
		baseURL:        s.URL,
		apiKey:         "some-api-key",
		batchSize:      1,
		uploadInterval: time.Duration(1 * time.Second),
		httpClient:     http.DefaultClient,
		taskManager:    &testWorkingTaskManager{},
		actionsBatch:   make([]ActionContainer, 0),
		tasks:          make(chan uploadTask),
		doneCh:         make(chan struct{}),
	}

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

	s := httptest.NewServer(&testAcceptingHandler{feedbackCh})
	defer s.Close()

	u := &uploaderImpl{
		baseURL:        s.URL,
		apiKey:         "some-api-key",
		batchSize:      1,
		uploadInterval: time.Duration(1 * time.Second),
		httpClient:     http.DefaultClient,
		taskManager:    &testWorkingTaskManager{},
		actionsBatch:   make([]ActionContainer, 0),
		tasks:          make(chan uploadTask),
		doneCh:         make(chan struct{}),
	}

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

	s := httptest.NewServer(&testAcceptingHandler{nil})
	defer s.Close()

	u := &uploaderImpl{
		baseURL:        s.URL,
		apiKey:         "some-api-key",
		batchSize:      1,
		uploadInterval: time.Duration(1 * time.Second),
		httpClient:     http.DefaultClient,
		taskManager:    &testWorkingTaskManager{},
		actionsBatch:   make([]ActionContainer, 0),
		tasks:          make(chan uploadTask),
		doneCh:         make(chan struct{}),
	}

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

	s := httptest.NewServer(&testAcceptingHandler{feedbackCh})
	defer s.Close()

	u := &uploaderImpl{
		baseURL:        s.URL,
		apiKey:         "some-api-key",
		batchSize:      100,
		uploadInterval: time.Duration(20 * time.Second),
		httpClient:     http.DefaultClient,
		taskManager:    &testWorkingTaskManager{},
		actionsBatch:   make([]ActionContainer, 0),
		tasks:          make(chan uploadTask),
		doneCh:         make(chan struct{}),
	}

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
