//go:build (linux || darwin || windows) && (amd64 || arm64)

package sqlite

import (
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/go-gorm/caches/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/backend/sqlite/tables"
	"github.com/xmapst/osreapi/internal/storage/models"
	"github.com/xmapst/osreapi/pkg/logx"
)

type Sqlite struct {
	*gorm.DB
}

func New(path string) (backend.IStorage, error) {
	option := "?cache=shared&mode=rwc&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	dialector := sqlite.Open(filepath.Join(path, "osreapi.db3"+option))
	gdb, err := gorm.Open(dialector, &gorm.Config{
		SkipDefaultTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:         "t_",
			SingularTable:       true,
			NoLowerCase:         false,
			IdentifierMaxLength: 256,
		},
		FullSaveAssociations: true,
		Logger: logger.New(logx.GetSubLogger(), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
			LogLevel:                  logger.Error,
		}),
		TranslateError: true,
	})
	if err != nil {
		return nil, err
	}
	db := &Sqlite{DB: gdb}
	err = db.init()
	return db, nil
}

func (s *Sqlite) init() error {
	// 开启外键约束
	s.Exec("PRAGMA foreign_keys=ON;")

	// 自动迁移表
	if err := s.AutoMigrate(
		&tables.Task{},
		&tables.TaskEnv{},
		&tables.Step{},
		&tables.StepEnv{},
		&tables.StepDepend{},
		&tables.Log{},
	); err != nil {
		logx.Errorln(err)
		return err
	}
	if err := s.Use(&caches.Caches{Conf: &caches.Config{
		Easer: true,
	}}); err != nil {
		return err
	}
	return nil
}

func (s *Sqlite) Name() string {
	return s.DB.Name()
}

func (s *Sqlite) Close() error {
	for {
		time.Sleep(100 * time.Millisecond)
		if tables.Rows() == 0 {
			break
		}
	}
	db, err := s.DB.DB()
	if err != nil {
		return err
	}
	defer logx.Infoln("close the database gracefully")
	return db.Close()
}

func (s *Sqlite) Task(name string) backend.ITask {
	return &task{
		db:   s.DB,
		name: name,
	}
}

func (s *Sqlite) TaskList(str string) (res []*models.Task) {
	if str != "" {
		s.Model(&tables.Task{}).Where("name LIKE ?", "%s"+str+"%").Order("s_time DESC, id DESC").Find(&res)
		return
	}
	s.Model(&tables.Task{}).Order("s_time DESC, id DESC").Find(&res)
	return
}

func (s *Sqlite) Step(taskName string, name string) backend.IStep {
	return &step{
		db:       s.DB,
		taskName: taskName,
		name:     name,
	}
}

func (s *Sqlite) Log(taskName string, stepName string) backend.ILog {
	return &log{
		db:       s.DB,
		taskName: taskName,
		stepName: stepName,
	}
}
