package types

type SPool struct {
	Size    int   `json:"size" form:"type" yaml:"Size" binding:"required" example:"30"`
	Total   int64 `json:"total" yaml:"Total"`
	Running int64 `json:"running" yaml:"Running"`
	Waiting int64 `json:"waiting" yaml:"Waiting"`
}
