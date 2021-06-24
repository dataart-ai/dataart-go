package task

import (
	"github.com/google/uuid"
)

type Task struct {
	uid  uuid.UUID
	Work func() error
}

func NewTask(function func() error) Task {
	return Task{
		uid:  uuid.New(),
		Work: function,
	}
}
