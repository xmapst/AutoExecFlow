package models

type Envs []*Env

type Env struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}
