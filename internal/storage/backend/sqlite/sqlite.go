package sqlite

import (
	"os"
	"path/filepath"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/xmapst/osreapi/internal/storage/backend"
	"github.com/xmapst/osreapi/internal/storage/backend/sqlite/tables"
	"github.com/xmapst/osreapi/internal/storage/models"
	"github.com/xmapst/osreapi/pkg/logx"
)

type storage struct {
	*gorm.DB
}

func New(path string) (backend.IStorage, error) {
	option := "?cache=shared&mode=rwc&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	dialector := sqlite.Open(filepath.Join(path, "osreapi.db3"+option))
	config := &gorm.Config{
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
	}
	db := new(storage)
	err := retry.Do(
		func() (err error) {
			db.DB, err = gorm.Open(dialector, config)
			if err != nil {
				// 尝试删除后重建
				_ = os.RemoveAll(path)
				_ = os.MkdirAll(path, os.ModeDir)
				return err
			}
			err = db.init()
			if err != nil {
				// 尝试删除后重建
				_ = os.RemoveAll(path)
				_ = os.MkdirAll(path, os.ModeDir)
				return err
			}
			return
		},
		retry.Attempts(3),
		retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
			_max := time.Duration(n)
			if _max > 8 {
				_max = 8
			}
			duration := time.Second * _max * _max
			return duration
		}),
	)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (s *storage) init() error {
	// 开启外键约束
	s.Exec("PRAGMA foreign_keys=ON;")
	s.Exec("PRAGMA sqlite_threadsafe=15000;")

	// 自动迁移表
	if err := s.AutoMigrate(
		&tables.Task{},
		&tables.TaskEnv{},
		&tables.Step{},
		&tables.StepEnv{},
		&tables.StepDepend{},
		&tables.StepLog{},
	); err != nil {
		logx.Errorln(err)
		return err
	}
	s.DB.Model(&tables.Task{}).Select("state = ?", models.Running).Updates(map[string]interface{}{
		"state":   models.Failed,
		"message": "unexpected ending",
	})
	s.DB.Model(&tables.Step{}).Select("state = ?", models.Running).Updates(map[string]interface{}{
		"state":   models.Failed,
		"message": "unexpected ending",
	})
	return nil
}

func (s *storage) Name() string {
	return s.DB.Name()
}

func (s *storage) Close() error {
	time.Sleep(3 * time.Second)
	return nil
}

func (s *storage) Task(name string) backend.ITask {
	return &task{
		db:    s.DB,
		tName: name,
	}
}

func (s *storage) TaskList(str string) (res models.Tasks) {
	if str != "" {
		s.Model(&tables.Task{}).Where("name LIKE ?", str+"%").Order("s_time DESC, id DESC").Find(&res)
		return
	}
	s.Model(&tables.Task{}).Order("s_time DESC, id DESC").Find(&res)
	return
}
