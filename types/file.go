package types

type SFileRes struct {
	Name    string `json:"name" yaml:"Name"`
	Path    string `json:"path" yaml:"Path"`
	Size    int64  `json:"size" yaml:"Size"`
	Mode    string `json:"mode" yaml:"Mode"`
	ModTime int64  `json:"mod_time" yaml:"ModTime"`
	IsDir   bool   `json:"is_dir" yaml:"IsDir"`
}

type SFileListRes struct {
	Total int       `json:"total" yaml:"Total"`
	Files SFilesRes `json:"files" yaml:"Files"`
}

type SFilesRes []*SFileRes
