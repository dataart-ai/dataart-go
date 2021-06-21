package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Uploader interface {
	UploadActions(cnt ActionsContainer) error
	UploadIdentity(cnt IdentityContainer) error
}

type uploaderImpl struct {
	apiKey     string
	httpClient *http.Client
}

func (u *uploaderImpl) upload(b []byte, url string) error {
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
		return fmt.Errorf("request failed with status code %d", res.StatusCode)
	}

	return nil
}

func (u *uploaderImpl) UploadActions(cnt ActionsContainer) error {
	b, err := json.Marshal(cnt)
	if err != nil {
		return err
	}

	return u.upload(b, actionsURL)
}

func (u *uploaderImpl) UploadIdentity(cnt IdentityContainer) error {
	b, err := json.Marshal(cnt)
	if err != nil {
		return err
	}

	return u.upload(b, identitiesURL)
}

func NewUploader(apiKey string, httpClient *http.Client) Uploader {
	return &uploaderImpl{
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}
