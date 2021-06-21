package dataart

import (
	"errors"
	"time"

	"github.com/dataart-ai/dataart-go/internal/http"
)

type Tracker interface {
	// EmitAction
	EmitAction(key string, userKey string, isAnonymousUser bool, timestamp time.Time, metadata map[string]interface{}) error
	// Identify
	Identify(userKey string, metadata map[string]interface{}) error
}

type trackerImpl struct {
	uploader http.Uploader
}

func (t *trackerImpl) EmitAction(key string, userKey string, isAnonymousUser bool, timestamp time.Time, metadata map[string]interface{}) error {
	if len(key) == 0 {
		return errors.New("event key identifier must not empty")
	}

	cnt := http.ActionsContainer{
		Timestamp: time.Now(),
		Actions: []http.Action{
			{
				Key:             key,
				UserKey:         userKey,
				IsAnonymousUser: isAnonymousUser,
				Timestamp:       timestamp,
				Metadata:        metadata,
			},
		},
	}

	return t.uploader.UploadActions(cnt)
}

func (t *trackerImpl) Identify(userKey string, metadata map[string]interface{}) error {
	if len(userKey) == 0 {
		return errors.New("userKey must not empty")
	}

	cnt := http.IdentityContainer{
		UserKey:  userKey,
		Metadata: metadata,
	}

	return t.uploader.UploadIdentity(cnt)
}
