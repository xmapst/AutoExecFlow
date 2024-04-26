package tables

import (
	"sync/atomic"
	"time"

	"gorm.io/gorm"

	"github.com/xmapst/osreapi/pkg/logx"
)

var rows int64

func Rows() int64 {
	return atomic.LoadInt64(&rows)
}

type Base struct {
	gorm.Model
}

type BaseModel interface {
	GetID() uint
	GetCreateTime() time.Time
	GetUpdateTime() time.Time
	GetDeleteTime() time.Time
}

func (b *Base) GetID() uint {
	return b.ID
}

func (b *Base) GetCreateTime() time.Time {
	return b.CreatedAt
}

func (b *Base) GetUpdateTime() time.Time {
	return b.UpdatedAt
}

func (b *Base) GetDeleteTime() time.Time {
	return b.DeletedAt.Time
}

func (b *Base) BeforeCreate(tx *gorm.DB) error {
	atomic.AddInt64(&rows, 1)
	return nil
}

func (b *Base) AfterCreate(tx *gorm.DB) error {
	atomic.AddInt64(&rows, -1)
	logx.Debugln(tx.Dialector.Explain(tx.Statement.SQL.String(), tx.Statement.Vars...))
	return nil
}

func (b *Base) BeforeUpdate(tx *gorm.DB) error {
	atomic.AddInt64(&rows, 1)
	return nil
}

func (b *Base) AfterUpdate(tx *gorm.DB) error {
	atomic.AddInt64(&rows, -1)
	logx.Debugln(tx.Dialector.Explain(tx.Statement.SQL.String(), tx.Statement.Vars...))
	return nil
}

func (b *Base) BeforeDelete(tx *gorm.DB) error {
	atomic.AddInt64(&rows, 1)
	return nil
}

func (b *Base) AfterDelete(tx *gorm.DB) error {
	atomic.AddInt64(&rows, -1)
	logx.Debugln(tx.Dialector.Explain(tx.Statement.SQL.String(), tx.Statement.Vars...))
	return nil
}
