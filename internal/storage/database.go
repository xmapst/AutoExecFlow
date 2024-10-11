package storage

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/xmapst/AutoExecFlow/internal/runner/common"
	"github.com/xmapst/AutoExecFlow/internal/storage/models"
	"github.com/xmapst/AutoExecFlow/internal/utils"
	"github.com/xmapst/AutoExecFlow/pkg/logx"
)

type database struct {
	*gorm.DB
}

func newDB(rawURL string) (*database, error) {
	before, after, found := strings.Cut(rawURL, "://")
	if !found {
		return nil, errors.New("invalid storage url")
	}
	var dialector gorm.Dialector
	switch before {
	case TYPE_MYSQL:
		dialector = mysql.Open(after)
	case TYPE_SQLITE:
		dialector = sqlite.Open(after)
	default:
		return nil, errors.New("unsupported storage type")
	}
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
	gdb, err := gorm.Open(dialector, config)
	if err != nil {
		return nil, err
	}
	_db, err := gdb.DB()
	if err != nil {
		return nil, err
	}
	//设置空闲连接池中连接的最大数量
	_db.SetMaxIdleConns(runtime.NumCPU())
	//设置打开数据库连接的最大数量
	_db.SetMaxIdleConns(runtime.NumCPU() * 15)

	d := &database{DB: gdb}

	if before == TYPE_SQLITE {
		d.initSqlite()
	}

	// 自动迁移表
	if err = d.AutoMigrate(
		&models.Task{},
		&models.TaskEnv{},
		&models.Step{},
		&models.StepEnv{},
		&models.StepDepend{},
		&models.StepLog{},
	); err != nil {
		logx.Errorln(err)
		return nil, err
	}

	// 查找当前节点的所有任务, 包括空的任务名称列表
	var tasks []string
	d.Model(&models.Task{}).
		Select("name").
		Where("(node IS NULL OR node = ?) AND (state <> ? AND state <> ?)", utils.HostName(), models.StateStopped, models.StateFailed).
		Find(&tasks)

	// 修正非正常关机时步骤还在运行中或挂起的状态为错误
	for _, taskName := range tasks {
		d.Model(&models.Task{}).
			Where("name = ?", taskName).
			Updates(map[string]interface{}{
				"node":    utils.HostName(),
				"state":   models.StateFailed,
				"message": "execution failed due to system error",
			})
		d.Model(&models.Step{}).
			Where("state = ? OR state = ?", models.StateRunning, models.StatePaused).
			Updates(map[string]interface{}{
				"state":   models.StateFailed,
				"code":    common.CodeSystemErr,
				"message": "execution failed due to system error",
			})
	}

	return d, nil
}

func (d *database) initSqlite() {
	// 开启外键约束
	d.Exec("PRAGMA foreign_keys=ON;")
	// 写同步
	d.Exec("PRAGMA synchronous=FULL;")
	// sqlite线程数
	d.Exec(fmt.Sprintf("PRAGMA sqlite_threadsafe=%d;", runtime.NumCPU()*15))
	// 启用 WAL 模式
	d.Exec("PRAGMA journal_mode=WAL;")
	d.Exec("PRAGMA journal_size_limit=104857600;")
	d.Exec("PRAGMA busy_timeout=999999;")
	d.Exec("PRAGMA cache=shared;")
	d.Exec("PRAGMA mode=rwc;")
}

func (d *database) Name() string {
	return d.DB.Name()
}

func (d *database) Close() error {
	db, err := d.DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

func (d *database) Task(name string) ITask {
	return &task{
		DB:    d.DB,
		tName: name,
	}
}

func (d *database) TaskCreate(task *models.Task) (err error) {
	return d.Create(task).Error
}

func (d *database) TaskCount(state models.State) (res int64) {
	if state != models.StateAll {
		d.Model(&models.Task{}).Distinct("DISTINCT name").Where("state = ?", state).Count(&res)
		return
	}
	d.Model(&models.Task{}).Distinct("DISTINCT name").Count(&res)
	return
}

func (d *database) TaskList(page, pageSize int64, str string) (res models.Tasks, total int64) {
	query := d.Model(&models.Task{}).
		Order("id DESC")
	if str != "" {
		query.Where("name LIKE ?", str+"%")
	}
	query.Count(&total).
		Scopes(func(db *gorm.DB) *gorm.DB {
			return models.Paginate(db, page, pageSize)
		}).Find(&res)
	return
}
