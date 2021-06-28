package http

import (
	"net/url"
	"path"
)

func buildURL(baseURL, endpointURL string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, endpointURL)
	return u.String(), nil
}

func buildActionsURL(baseURL string) (string, error) {
	return buildURL(baseURL, "/events/send-actions")
}

func buildIdentitiesURL(baseURL string) (string, error) {
	return buildURL(baseURL, "/users/identify")
}
