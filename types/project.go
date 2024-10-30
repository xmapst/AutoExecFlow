package types

type SProjectListDetailRes struct {
	Page     SPageRes        `json:"page" yaml:"Page"`
	Projects SProjectListRes `json:"projects" yaml:"Projects"`
}

type SProjectListRes []*SProjectRes

type SProjectRes struct {
	Name    string `json:"name" yaml:"Name"`
	Disable bool   `json:"disable" yaml:"Disable"`
}
