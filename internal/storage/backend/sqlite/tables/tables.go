package tables

import (
	"github.com/xmapst/osreapi/internal/storage/models"
)

type Task struct {
	Base
	models.Task
}

type TaskEnv struct {
	Base
	models.TaskEnv
}

type Step struct {
	Base
	models.Step
}

type StepEnv struct {
	Base
	models.StepEnv
}

type StepDepend struct {
	Base
	models.StepDepend
}

type Log struct {
	Base
	models.Log
}
