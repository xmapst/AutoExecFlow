package models

type Envs []*Env

type Env struct {
	Name  string `json:"name,omitempty" gorm:"not null;comment:名称"`
	Value string `json:"value,omitempty" gorm:"comment:值"`
}
