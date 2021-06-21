package http

import "time"

type Action struct {
	Key             string                 `json:"key"`
	UserKey         string                 `json:"user_key"`
	IsAnonymousUser bool                   `json:"is_anonymous_user"`
	Timestamp       time.Time              `json:"timestamp"`
	Metadata        map[string]interface{} `json:"metadata"`
}

type ActionsContainer struct {
	Timestamp time.Time `json:"timestamp"`
	Actions   []Action  `json:"actions"`
}

type IdentityContainer struct {
	UserKey  string                 `json:"user_key"`
	Metadata map[string]interface{} `json:"metadata"`
}
