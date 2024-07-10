package types

type FileRes struct {
	URL     string `json:"url" yaml:"URL"`
	Name    string `json:"name" yaml:"Name"`
	Path    string `json:"path" yaml:"Path"`
	Size    int64  `json:"size" yaml:"Size"`
	Mode    string `json:"mode" yaml:"Mode"`
	ModTime int64  `json:"mod_time" yaml:"ModTime"`
	IsDir   bool   `json:"is_dir" yaml:"IsDir"`
}

type FileListRes struct {
	Total int        `json:"total" yaml:"Total"`
	Files []*FileRes `json:"files" yaml:"Files"`
}
