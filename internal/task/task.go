package task

import (
	"github.com/dataart-ai/dataart-go/internal/pkg/randomutil"
)

const (
	taskIDLength = 16
)

type Task struct {
	id   string
	Work func() error
}

func NewTask(function func() error) Task {
	return Task{
		id:   randomutil.String(taskIDLength),
		Work: function,
	}
}
