package task

import (
	"github.com/dataart-ai/dataart-go/internal/pkg/randomutil"
)

const (
	taskIDLength = 16
)

type task struct {
	id   string
	work func() error
}

func newTask(work func() error) task {
	return task{
		id:   randomutil.String(taskIDLength),
		work: work,
	}
}
