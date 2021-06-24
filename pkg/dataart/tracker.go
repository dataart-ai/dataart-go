package dataart

import (
	"errors"
	"time"

	"github.com/dataart-ai/dataart-go/internal/http"
)

type Tracker interface {
	// EmitAction creates an action object with given properties and hands it to uploader.
	EmitAction(key string, userKey string, isAnonymousUser bool, timestamp time.Time, metadata map[string]interface{}) error

	// Identify creates an identity object with given properties and hands it to uploader.
	Identify(userKey string, metadata map[string]interface{}) error

	// Close gracefully shuts down the underlying uploader instance.
	Close()
}

type trackerImpl struct {
	uploader http.Uploader
}

func (t *trackerImpl) EmitAction(key string, userKey string, isAnonymousUser bool,
	timestamp time.Time, metadata map[string]interface{}) error {

	if len(key) == 0 {
		return errors.New("event key identifier must not empty")
	}

	return t.uploader.UploadAction(
		http.ActionContainer{
			Key:             key,
			UserKey:         userKey,
			IsAnonymousUser: isAnonymousUser,
			Timestamp:       timestamp,
			Metadata:        metadata,
		},
	)
}

func (t *trackerImpl) Identify(userKey string, metadata map[string]interface{}) error {
	if len(userKey) == 0 {
		return errors.New("userKey must not empty")
	}

	return t.uploader.UploadIdentity(
		http.IdentityContainer{
			UserKey:  userKey,
			Metadata: metadata,
		},
	)
}

func (t *trackerImpl) Close() {
	t.uploader.Shutdown()
}
