package bbolt

import (
	"go.etcd.io/bbolt"

	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/models"
)

type task struct {
	db   *bbolt.DB
	name string
}

func (t *task) ClearAll() {
	//TODO implement me
	panic("implement me")
}

func (t *task) Delete() (err error) {
	//TODO implement me
	panic("implement me")
}

func (t *task) GetState() (state int, err error) {
	//TODO implement me
	panic("implement me")
}

func (t *task) Env() backend.IEnv {
	return &taskEnv{
		db:       t.db,
		taskName: t.name,
	}
}

func (t *task) Get() (res *models.Task, err error) {
	//TODO implement me
	panic("implement me")
}

func (t *task) Create(task *models.Task) (err error) {
	//TODO implement me
	panic("implement me")
}

func (t *task) Update(value *models.TaskUpdate) (err error) {
	//TODO implement me
	panic("implement me")
}

func (t *task) Step(name string) backend.IStep {
	return &step{
		db:       t.db,
		taskName: t.name,
		name:     name,
	}
}

func (t *task) StepList(str string) (res []*models.Step) {
	//TODO implement me
	panic("implement me")
}

type taskEnv struct {
	db       *bbolt.DB
	taskName string
}

func (t *taskEnv) List() (res []*models.Env) {
	//TODO implement me
	panic("implement me")
}

func (t *taskEnv) Create(env []*models.Env) (err error) {
	//TODO implement me
	panic("implement me")
}

func (t *taskEnv) Get(name string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (t *taskEnv) Delete(name string) (err error) {
	//TODO implement me
	panic("implement me")
}

func (t *taskEnv) DeleteAll() (err error) {
	//TODO implement me
	panic("implement me")
}
