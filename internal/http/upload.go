package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/dataart-ai/dataart-go/internal/pkg/atomicutil"
	"github.com/dataart-ai/dataart-go/internal/task"
)

const (
	objTypeAction   = "action"
	objTypeIdentity = "identity"
)

type Uploader interface {
	UploadAction(cnt ActionContainer) error
	UploadIdentity(cnt IdentityContainer) error
	Shutdown()
}

type uploadTask struct {
	objType string
	obj     interface{}
}

type uploaderImpl struct {
	baseURL        string
	apiKey         string
	batchSize      int
	uploadInterval time.Duration
	httpClient     *http.Client

	taskManager task.Manager

	actionsBatch []ActionContainer

	tasks         chan uploadTask
	doneCh        chan struct{}
	debugTickerCh chan bool

	once       sync.Once
	inShutdown atomicutil.Bool
	isStarted  atomicutil.Bool
}

func (u *uploaderImpl) buildRequest(url string, b []byte) func() error {
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

func (u *uploaderImpl) flushIdentity(cnt IdentityContainer) {
	b, _ := json.Marshal(cnt)

	u.taskManager.Queue(
		task.NewTask(u.buildRequest(buildIdentitiesURL(u.baseURL), b)),
	)
}

func (u *uploaderImpl) flushActions() {
	dup := make([]ActionContainer, len(u.actionsBatch))
	copy(dup, u.actionsBatch)

	cnt := ActionsContainer{
		Timestamp: time.Now(),
		Actions:   dup,
	}

	b, _ := json.Marshal(cnt)

	u.taskManager.Queue(
		task.NewTask(u.buildRequest(buildActionsURL(u.baseURL), b)),
	)

	u.actionsBatch = make([]ActionContainer, 0)
}

func (u *uploaderImpl) start() {
	u.isStarted.SetTrue()

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
				if u.debugTickerCh != nil {
					u.debugTickerCh <- true
				}
				if len(u.actionsBatch) > 0 {
					u.flushActions()
				}
			case <-u.doneCh:
				t.Stop()
				if len(u.actionsBatch) > 0 {
					u.flushActions()
				}
				return
			}
		}
	}()
}

func (u *uploaderImpl) UploadAction(cnt ActionContainer) error {
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

func (u *uploaderImpl) UploadIdentity(cnt IdentityContainer) error {
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

func (u *uploaderImpl) Shutdown() {
	if u.inShutdown.IsSet() || !u.isStarted.IsSet() {
		return
	}
	u.inShutdown.SetTrue()

	u.doneCh <- struct{}{}
	u.taskManager.Shutdown()
}

func NewUploader(baseURL string, apiKey string, batchSize int, uploadInterval time.Duration,
	httpClient *http.Client, taskManager task.Manager) (Uploader, error) {

	if len(baseURL) == 0 {
		return nil, errors.New("baseURL must not be empty")
	}

	if len(apiKey) == 0 {
		return nil, errors.New("apiKey must not be empty")
	}

	if batchSize < 1 {
		return nil, errors.New("batchSize must be at least 1")
	}

	if uploadInterval < time.Duration(1*time.Second) {
		return nil, errors.New("uploadInterval can't be less 1 second")
	}

	if httpClient == nil {
		return nil, errors.New("httpClient can't be nil")
	}

	if taskManager == nil {
		return nil, errors.New("taskManager can't be nil")
	}

	u := &uploaderImpl{
		baseURL:        baseURL,
		apiKey:         apiKey,
		batchSize:      batchSize,
		uploadInterval: uploadInterval,
		httpClient:     httpClient,
		taskManager:    taskManager,
		actionsBatch:   make([]ActionContainer, 0),
		tasks:          make(chan uploadTask),
		doneCh:         make(chan struct{}),
		debugTickerCh:  nil,
	}

	return u, nil
}
