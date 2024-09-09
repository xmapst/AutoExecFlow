package types

type Pool struct {
	Size    int   `json:"size" form:"type" yaml:"Size" binding:"required" example:"30"`
	Total   int64 `json:"total" yaml:"Total" swaggerignore:"true"`
	Running int64 `json:"running" yaml:"Running" swaggerignore:"true"`
	Waiting int64 `json:"waiting" yaml:"Waiting" swaggerignore:"true"`
}
