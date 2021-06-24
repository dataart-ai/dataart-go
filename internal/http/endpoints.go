package http

func buildActionsURL(baseURL string) string {
	return baseURL + "/events/send-actions"
}

func buildIdentitiesURL(baseURL string) string {
	return baseURL + "/users/identify"
}
