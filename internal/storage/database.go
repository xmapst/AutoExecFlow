package storage

import (
	"runtime/debug"
	"strings"
	"time"

	"github.com/go-gorm/caches/v4"
	"github.com/pkg/errors"
	"github.com/xmapst/logx"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/xmapst/AutoExecFlow/internal/storage/models"
	"github.com/xmapst/AutoExecFlow/internal/worker/common"
)

type sDatabase struct {
	*gorm.DB
}

func newDB(rawURL string) (*sDatabase, error) {
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
		NamingStrategy: schema.NamingStrategy{
			SingularTable:       true,
			NoLowerCase:         false,
			IdentifierMaxLength: 256,
		},
		Logger: logger.New(logx.GetSubLogger(), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			Colorful:                  false,
			IgnoreRecordNotFoundError: true,
			LogLevel:                  logger.Warn,
		}),
		SkipDefaultTransaction: true,
		FullSaveAssociations:   true,
		TranslateError:         true,
	}
	gdb, err := gorm.Open(dialector, config)
	if err != nil {
		return nil, err
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	_ = gdb.Use(&caches.Caches{Conf: &caches.Config{
		Easer: true,
	}})

	d := &sDatabase{DB: gdb}

	if before == TYPE_SQLITE {
		d.initSqlite()
	}

	// 自动迁移表
	if err = d.AutoMigrate(
		&models.STask{},
		&models.STaskEnv{},
		&models.SStep{},
		&models.SStepEnv{},
		&models.SStepDepend{},
		&models.SStepLog{},
		&models.SPipeline{},
		&models.SPipelineBuild{},
	); err != nil {
		logx.Errorln(err)
		return nil, err
	}

	return d, nil
}

func (d *sDatabase) initSqlite() {
	// 开启外键约束
	d.Exec("PRAGMA foreign_keys=ON;")
	// 写同步
	d.Exec("PRAGMA synchronous=NORMAL;")
	// 启用 WAL 模式
	d.Exec("PRAGMA journal_mode=WAL;")
	// 控制WAL文件大小 100MB
	d.Exec("PRAGMA journal_size_limit=104857600;")
	// 设置等待超时，减少锁等待时间 5秒
	d.Exec("PRAGMA busy_timeout=5000;")
	// 设置共享缓存
	d.Exec("PRAGMA cache=shared;")
	// 设置缓存大小 约32MB缓存
	d.Exec("PRAGMA cache_size=-8000;")
	// 设置内存映射大小 128MB
	d.Exec("PRAGMA mmap_size=134217728;")
	// 将临时表放入内存
	d.Exec("PRAGMA temp_store=MEMORY;")
	// 设置锁模式为NORMAL，支持高并发访问
	d.Exec("PRAGMA locking_mode=NORMAL;")
	// 开启缓存溢出管理，适用于高并发写入
	d.Exec("PRAGMA cache_spill=ON;")
}

func (d *sDatabase) Name() string {
	return d.DB.Name()
}

func (d *sDatabase) Close() error {
	db, err := d.DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

func (d *sDatabase) GetDB() *gorm.DB {
	return d.DB
}

func (d *sDatabase) FixDatabase(nodeName string) (err error) {
	// 开始事务
	tx := d.Begin()
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			logx.Errorln(r, string(stack))
			tx.Rollback()
		}
	}()

	// 更新所有符合条件的任务状态为失败
	if err = tx.Model(&models.STask{}).
		Where("(node IS NULL OR node = ?) AND (state <> ? AND state <> ? AND state <> ?)", nodeName, models.StateStopped, models.StateSkipped, models.StateFailed).
		Updates(map[string]interface{}{
			"node":    nodeName,
			"state":   models.StateFailed,
			"message": "execution failed due to system error",
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新所有符合条件的步骤状态为失败
	if err = tx.Model(&models.SStep{}).
		Where("task_name IN (?)",
			d.Model(&models.STask{}).Select("name").
				Where("(node IS NULL OR node = ?) AND (state <> ? AND state <> ? AND state <> ?)", nodeName, models.StateStopped, models.StateSkipped, models.StateFailed),
		).
		Where("state = ? OR state = ?", models.StateRunning, models.StatePaused).
		Updates(map[string]interface{}{
			"state":   models.StateFailed,
			"code":    common.CodeSystemErr,
			"message": "execution failed due to system error",
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 提交事务
	if err = tx.Commit().Error; err != nil {
		return err
	}
	return
}

func (d *sDatabase) Task(name string) ITask {
	return &sTask{
		DB:    d.DB,
		tName: name,
	}
}

func (d *sDatabase) TaskCreate(task *models.STask) (err error) {
	return d.Create(task).Error
}

func (d *sDatabase) TaskCount(state models.State) (res int64) {
	if state != models.StateAll {
		d.Model(&models.STask{}).Distinct("DISTINCT name").Where("state = ?", state).Count(&res)
		return
	}
	d.Model(&models.STask{}).Distinct("DISTINCT name").Count(&res)
	return
}

func (d *sDatabase) TaskList(page, pageSize int64, str string) (res models.STasks, total int64) {
	err := d.Model(&models.STask{}).Count(&total).Error
	if err != nil {
		return
	}
	query := d.Model(&models.STask{}).
		Select("id, name, state, message, s_time, e_time").
		Order("id DESC")
	if str != "" {
		query.Where("name LIKE ?", str+"%")
	}
	query.Scopes(func(db *gorm.DB) *gorm.DB {
		return models.Paginate(db, page, pageSize)
	}).Find(&res)
	return
}

func (d *sDatabase) Pipeline(name string) IPipeline {
	return &sPipeline{
		DB:   d.DB,
		name: name,
	}
}

func (d *sDatabase) PipelineCreate(pipeline *models.SPipeline) (err error) {
	return d.Create(pipeline).Error
}

func (d *sDatabase) PipelineList(page, pageSize int64, str string) (res models.SPipelines, total int64) {
	err := d.Model(&models.SPipeline{}).Count(&total).Error
	if err != nil {
		return
	}
	query := d.Model(&models.SPipeline{}).
		Select("id, name, disable, tpl_type").
		Order("id DESC")
	if str != "" {
		query.Where("name LIKE ?", str+"%")
	}
	query.Scopes(func(db *gorm.DB) *gorm.DB {
		return models.Paginate(db, page, pageSize)
	}).Find(&res)
	return
}
