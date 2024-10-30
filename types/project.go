package types

type SProjectListRes struct {
	Page     SPageRes       `json:"page" yaml:"Page"`
	Projects []*SProjectRes `json:"projects" yaml:"Projects"`
}

type SProjectRes struct {
	Name    string `json:"name" yaml:"Name"`
	Disable bool   `json:"disable" yaml:"Disable"`
}
