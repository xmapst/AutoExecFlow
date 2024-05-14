package models

import (
	"time"
)

type Step struct {
	Name    string        `json:"name,omitempty"`
	Type    string        `json:"type,omitempty"`
	Content string        `json:"content,omitempty"`
	Timeout time.Duration `json:"timeout,omitempty"`
	StepUpdate
}

type StepUpdate struct {
	Message  string     `json:"message,omitempty"`
	State    *int       `json:"state,omitempty"`
	OldState *int       `json:"old_state,omitempty"`
	Code     *int64     `json:"code,omitempty"`
	STime    *time.Time `json:"s_time,omitempty"`
	ETime    *time.Time `json:"e_time,omitempty"`
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
