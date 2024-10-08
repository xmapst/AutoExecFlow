package sqlite

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/avast/retry-go/v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/xmapst/AutoExecFlow/internal/runner/common"
	"github.com/xmapst/AutoExecFlow/internal/storage/backend"
	"github.com/xmapst/AutoExecFlow/internal/storage/backend/sqlite/tables"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

type storage struct {
	*gorm.DB
}

func New(path string) (backend.IStorage, error) {
	dialector := sqlite.Open(filepath.Join(path, fmt.Sprintf("%s.db3", utils.ServiceName)))
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
	s := new(storage)
	err := retry.Do(
		func() (err error) {
			defer func() {
				if err == nil {
					return
				}
				// 尝试删除后重建
				_ = os.RemoveAll(path)
				_ = os.MkdirAll(path, os.ModePerm)
			}()
			s.DB, err = gorm.Open(dialector, config)
			if err != nil {
				return err
			}
			err = s.init()
			if err != nil {
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
	return s, nil
}

func (s *storage) init() error {
	db, err := s.DB.DB()
	if err != nil {
		return err
	}
	//设置空闲连接池中连接的最大数量
	db.SetMaxIdleConns(runtime.NumCPU())
	//设置打开数据库连接的最大数量
	db.SetMaxIdleConns(runtime.NumCPU() * 15)
	// 开启外键约束
	s.Exec("PRAGMA foreign_keys=ON;")
	// 写同步
	s.Exec("PRAGMA synchronous=FULL;")
	// sqlite线程数
	s.Exec(fmt.Sprintf("PRAGMA sqlite_threadsafe=%d;", runtime.NumCPU()*15))
	// 启用 WAL 模式
	s.Exec("PRAGMA journal_mode=WAL;")
	s.Exec("PRAGMA journal_size_limit=104857600;")
	s.Exec("PRAGMA busy_timeout=999999;")
	s.Exec("PRAGMA cache=shared;")
	s.Exec("PRAGMA mode=rwc;")

	// 自动迁移表
	if err = s.AutoMigrate(
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
	// 修正非正常关机时任务状态不为停止和失败的状态强制为错误
	s.Model(&tables.Task{}).
		Where("state <> ? AND state <> ?", models.StateStop, models.StateFailed).
		Updates(map[string]interface{}{
			"state":   models.StateFailed,
			"message": "execution failed due to system error",
		})
	// 修正非正常关机时步骤还在运行中或挂起的状态为错误
	s.Model(&tables.Step{}).
		Where("state = ? OR state = ?", models.StateRunning, models.StatePaused).
		Updates(map[string]interface{}{
			"state":   models.StateFailed,
			"code":    common.CodeSystemErr,
			"message": "execution failed due to system error",
		})
	return nil
}

func (s *storage) Name() string {
	return s.DB.Name()
}

func (s *storage) Close() error {
	db, err := s.DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

func (s *storage) Task(name string) backend.ITask {
	return &task{
		DB:    s.DB,
		tName: name,
	}
}

func (s *storage) TaskCount(state models.State) (res int64) {
	if state != models.StateAll {
		s.Model(&tables.Task{}).Distinct("DISTINCT name").Where("state = ?", state).Count(&res)
		return
	}
	s.Model(&tables.Task{}).Distinct("DISTINCT name").Count(&res)
	return
}

func (s *storage) TaskList(page, pageSize int64, str string) (res models.Tasks, total int64) {
	query := s.Model(&tables.Task{}).
		Order("id DESC")
	if str != "" {
		query.Where("name LIKE ?", str+"%")
	}
	query.Count(&total).
		Scopes(func(db *gorm.DB) *gorm.DB {
			return tables.Paginate(db, page, pageSize)
		}).Find(&res)
	return
}
