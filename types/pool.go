package types

type SPoolReq struct {
	Size int `json:"size" form:"type" yaml:"Size" binding:"required"`
}

type SPoolRes struct {
	Size    int   `json:"size" yaml:"Size"`
	Total   int64 `json:"total" yaml:"Total"`
	Running int64 `json:"running" yaml:"Running"`
	Waiting int64 `json:"waiting" yaml:"Waiting"`
}
