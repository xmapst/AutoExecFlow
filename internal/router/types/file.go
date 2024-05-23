package types

type FileRes struct {
	URL     string `json:"url"`
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	ModTime int64  `json:"mod_time"`
	IsDir   bool   `json:"is_dir"`
}

type FileListRes struct {
	Total int        `json:"total"`
	Files []*FileRes `json:"files"`
}
