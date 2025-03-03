package storage

import (
	"gorm.io/gorm"

	"github.com/xmapst/AutoExecFlow/internal/storage/models"
	"github.com/xmapst/AutoExecFlow/pkg/snowflake"
)

var storage IStorage

const (
	TypeSqlite    = "sqlite"
	TypeMysql     = "mysql"
	TypePostgres  = "postgres"
	TypeSqlserver = "sqlserver"
)

func New(dataCenterID, nodeID int64, rawURL string) error {
	db, err := newDB(rawURL)
	if err != nil {
		return err
	}
	storage = db
	_, err = snowflake.New(dataCenterID, nodeID)
	if err != nil {
		return err
	}
	models.DataCenterID = dataCenterID
	models.NodeID = nodeID
	return nil
}

func Name() string {
	return storage.Name()
}

func Close() error {
	return storage.Close()
}

func GetDB() *gorm.DB {
	return storage.GetDB()
}

func FixDatabase(nodeName string) (err error) {
	return storage.FixDatabase(nodeName)
}

func Task(name string) ITask {
	return storage.Task(name)
}

func TaskCreate(task *models.STask) (err error) {
	return storage.TaskCreate(task)
}

func TaskCount(state models.State) (res int64) {
	return storage.TaskCount(state)
}

func TaskList(page, pageSize int64, str string) (res []*models.STask, total int64) {
	return storage.TaskList(page, pageSize, str)
}

func Pipeline(name string) IPipeline {
	return storage.Pipeline(name)
}

func PipelineCreate(pipeline *models.SPipeline) (err error) {
	return storage.PipelineCreate(pipeline)
}

func PipelineList(page, pageSize int64, str string) (res []*models.SPipeline, total int64) {
	return storage.PipelineList(page, pageSize, str)
}
