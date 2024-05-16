package models

import (
	"time"
)

type Step struct {
	Name    string        `json:"name,omitempty" gorm:"not null;comment:名称"`
	Type    string        `json:"type,omitempty" gorm:"not null;comment:类型"`
	Content string        `json:"content,omitempty" gorm:"comment:内容"`
	Timeout time.Duration `json:"timeout,omitempty" gorm:"not null;default:86400000000000;comment:超时时间"`
	StepUpdate
}

type StepUpdate struct {
	Message  string     `json:"message,omitempty" gorm:"comment:消息"`
	State    *int       `json:"state,omitempty" gorm:"not null;default:0;comment:状态"`
	OldState *int       `json:"old_state,omitempty" gorm:"not null;default:0;comment:旧状态"`
	Code     *int64     `json:"code,omitempty" gorm:"not null;default:0;comment:退出码"`
	STime    *time.Time `json:"s_time,omitempty" gorm:"comment:开始时间"`
	ETime    *time.Time `json:"e_time,omitempty" gorm:"comment:结束时间"`
}

func (s *StepUpdate) STimeStr() string {
	if s.STime == nil {
		return ""
	}
	return s.STime.Format(time.RFC3339)
}

func (s *StepUpdate) ETimeStr() string {
	if s.ETime == nil {
		return ""
	}
	return s.ETime.Format(time.RFC3339)
}

type Steps []*Step

func (s Steps) Len() int {
	return len(s)
}

func (s Steps) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (s Steps) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
