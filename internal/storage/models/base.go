package models

import (
	"time"

	"gorm.io/gorm"
)

type State int

const (
	StateStopped State = iota // 成功
	StateRunning              // 运行
	StateFailed               // 失败
	StateUnknown              // 未知
	StatePending              // 等待
	StatePaused               // 挂起
	StateAll     State = -1
)

var StateMap = map[State]string{
	StateStopped: "stopped",
	StateRunning: "running",
	StateFailed:  "failed",
	StateUnknown: "unknown",
	StatePending: "pending",
	StatePaused:  "paused",
}

func Pointer[T any](v T) *T {
	return &v
}

type Base struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BaseModel interface {
	GetID() uint
	GetCreateTime() time.Time
	GetUpdateTime() time.Time
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

func (b *Base) BeforeCreate(tx *gorm.DB) error {
	return nil
}

func (b *Base) AfterCreate(tx *gorm.DB) error {
	//logx.Debugln(tx.Dialector.Explain(tx.Statement.SQL.String(), tx.Statement.Vars...))
	return nil
}

func (b *Base) BeforeUpdate(tx *gorm.DB) error {
	return nil
}

func (b *Base) AfterUpdate(tx *gorm.DB) error {
	//logx.Debugln(tx.Dialector.Explain(tx.Statement.SQL.String(), tx.Statement.Vars...))
	return nil
}

func (b *Base) BeforeDelete(tx *gorm.DB) error {
	return nil
}

func (b *Base) AfterDelete(tx *gorm.DB) error {
	//logx.Debugln(tx.Dialector.Explain(tx.Statement.SQL.String(), tx.Statement.Vars...))
	return nil
}

func Paginate(db *gorm.DB, page, pageSize int64) *gorm.DB {
	if page == -1 {
		return db
	}
	if page == 0 {
		page = 1
	}
	switch {
	case pageSize > 500:
		pageSize = 500
	case pageSize <= 0:
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	return db.Offset(int(offset)).Limit(int(pageSize))
}

type Envs []*Env

type Env struct {
	Name  string `json:"name,omitempty" gorm:"index;not null;comment:名称"`
	Value string `json:"value,omitempty" gorm:"comment:值"`
}
