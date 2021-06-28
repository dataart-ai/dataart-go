package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/dataart-ai/dataart-go/internal/pkg/atomicutil"
)

const (
	objTypeAction   = "action"
	objTypeIdentity = "identity"

	minUploadInterval = time.Duration(5 * time.Second)
)

type TaskManager interface {
	Queue(work func() error) error
	Shutdown()
}

type uploadTask struct {
	objType string
	obj     interface{}
}

// Uploader receives data objects and batches them if necessary in a request. These
// requests are then executed using a task manager.
type Uploader struct {
	baseURL        string
	apiKey         string
	batchSize      int
	uploadInterval time.Duration
	httpClient     *http.Client

	tasks  chan uploadTask
	doneCh chan struct{}

	actionsBatch []ActionContainer

	tm TaskManager

	wg         sync.WaitGroup
	once       sync.Once
	inShutdown atomicutil.Bool
	isStarted  atomicutil.Bool
}

func (u *Uploader) buildRequest(url string, b []byte) func() error {
	return func() error {
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
		if err != nil {
			return err
		}

		req.Header.Add("User-Agent", "dataart-go")
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Content-Length", fmt.Sprint(len(b)))
		req.Header.Add("X-API-Key", u.apiKey)

		res, err := u.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			content, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return fmt.Errorf(
					"request failed with status code %d. parsing response failed: %s", res.StatusCode, err.Error())
			}

			return fmt.Errorf(
				"request failed with status code %d. got response: %s", res.StatusCode, string(content))
		}

		return nil
	}
}

func (u *Uploader) flushIdentity(cnt IdentityContainer) {
	b, _ := json.Marshal(cnt)

	// Error checking is skipped since we validate baseURL in initialization.
	iurl, _ := buildIdentitiesURL(u.baseURL)

	u.tm.Queue(
		u.buildRequest(iurl, b),
	)
}

func (u *Uploader) flushActions() {
	dup := make([]ActionContainer, len(u.actionsBatch))
	copy(dup, u.actionsBatch)

	cnt := ActionsContainer{
		Timestamp: time.Now(),
		Actions:   dup,
	}

	b, _ := json.Marshal(cnt)

	// Error checking is skipped since we validate baseURL in initialization.
	aurl, _ := buildActionsURL(u.baseURL)

	u.tm.Queue(
		u.buildRequest(aurl, b),
	)

	u.actionsBatch = make([]ActionContainer, 0)
}

func (u *Uploader) start() {
	u.isStarted.SetTrue()

	u.wg.Add(1)
	go func() {
		for {
			t := time.NewTimer(u.uploadInterval)
			select {
			case t := <-u.tasks:
				switch t.objType {
				case objTypeAction:
					obj := t.obj.(ActionContainer)
					u.actionsBatch = append(u.actionsBatch, obj)
					if len(u.actionsBatch) == u.batchSize {
						u.flushActions()
					}
				case objTypeIdentity:
					obj := t.obj.(IdentityContainer)
					u.flushIdentity(obj)
				}
			case <-t.C:
				if len(u.actionsBatch) > 0 {
					u.flushActions()
				}
			case <-u.doneCh:
				t.Stop()
				if len(u.actionsBatch) > 0 {
					u.flushActions()
				}
				u.tm.Shutdown()
				u.wg.Done()
				return
			}
		}
	}()
}

// UploadAction queues given action object to be uploaded to server.
func (u *Uploader) UploadAction(cnt ActionContainer) error {
	if u.inShutdown.IsSet() {
		return errors.New("uploader is shutting down")
	}

	u.once.Do(u.start)

	t := uploadTask{
		objType: objTypeAction,
		obj:     cnt,
	}

	u.tasks <- t
	return nil
}

// UploadIdentity queues given identity object to be uploaded to server.
func (u *Uploader) UploadIdentity(cnt IdentityContainer) error {
	if u.inShutdown.IsSet() {
		return errors.New("uploader is shutting down")
	}

	u.once.Do(u.start)

	t := uploadTask{
		objType: objTypeIdentity,
		obj:     cnt,
	}

	u.tasks <- t
	return nil
}

// Shutdown terminates Uploader gracefully. It will flush all requests before
// closing the buffer and then returns.
func (u *Uploader) Shutdown() {
	if u.inShutdown.IsSet() || !u.isStarted.IsSet() {
		return
	}
	u.inShutdown.SetTrue()

	u.doneCh <- struct{}{}
	u.wg.Wait()
}

// NewUploader creates a new Uploader instance using provided values. Use this
// function to instantiate a concrete Uploader type.
func NewUploader(baseURL string, apiKey string, batchSize int, uploadInterval time.Duration,
	httpClient *http.Client, tm TaskManager) (*Uploader, error) {

	_, err := url.Parse(baseURL)
	if len(baseURL) == 0 || err != nil {
		return nil, errors.New("baseURL is not valid")
	}

	if len(apiKey) == 0 {
		return nil, errors.New("apiKey must not be empty")
	}

	if batchSize < 1 {
		return nil, errors.New("batchSize must be at least 1")
	}

	if uploadInterval < time.Duration(minUploadInterval) {
		return nil, errors.New("uploadInterval can't be less than 5 seconds")
	}

	if httpClient == nil {
		return nil, errors.New("httpClient can't be nil")
	}

	if tm == nil {
		return nil, errors.New("taskManager can't be nil")
	}

	u := &Uploader{
		baseURL:        baseURL,
		apiKey:         apiKey,
		batchSize:      batchSize,
		uploadInterval: uploadInterval,
		httpClient:     httpClient,
		tm:             tm,
		actionsBatch:   make([]ActionContainer, 0),
		tasks:          make(chan uploadTask),
		doneCh:         make(chan struct{}),
	}

	return u, nil
}
