package bbolt

import (
	"go.etcd.io/bbolt"

	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/models"
)

type step struct {
	db       *bbolt.DB
	taskName string
	name     string
}

func (s *step) ClearAll() {
	//TODO implement me
	panic("implement me")
}

func (s *step) Delete() (err error) {
	//TODO implement me
	panic("implement me")
}

func (s *step) GetState() (state int, err error) {
	//TODO implement me
	panic("implement me")
}

func (s *step) Env() backend.IEnv {
	return &stepEnv{
		db:       s.db,
		taskName: s.taskName,
		stepName: s.name,
	}
}

func (s *step) Get() (res *models.Step, err error) {
	//TODO implement me
	panic("implement me")
}

func (s *step) Create(step *models.Step) (err error) {
	//TODO implement me
	panic("implement me")
}

func (s *step) Update(value *models.StepUpdate) (err error) {
	//TODO implement me
	panic("implement me")
}

func (s *step) Depend() backend.IDepend {
	return &stepDepend{
		db:       s.db,
		taskName: s.taskName,
		stepName: s.name,
	}
}

func (s *step) Log() backend.ILog {
	return &log{
		db:       s.db,
		taskName: s.taskName,
		stepName: s.name,
	}
}

type stepEnv struct {
	db       *bbolt.DB
	taskName string
	stepName string
}

func (s *stepEnv) List() (res []*models.Env) {
	//TODO implement me
	panic("implement me")
}

func (s *stepEnv) Create(env []*models.Env) (err error) {
	//TODO implement me
	panic("implement me")
}

func (s *stepEnv) Get(name string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (s *stepEnv) Delete(name string) (err error) {
	//TODO implement me
	panic("implement me")
}

func (s *stepEnv) DeleteAll() (err error) {
	//TODO implement me
	panic("implement me")
}

type stepDepend struct {
	db       *bbolt.DB
	taskName string
	stepName string
}

func (s *stepDepend) List() (res []string) {
	//TODO implement me
	panic("implement me")
}

func (s *stepDepend) Create(depends []string) (err error) {
	//TODO implement me
	panic("implement me")
}

func (s *stepDepend) DeleteAll() (err error) {
	//TODO implement me
	panic("implement me")
}
