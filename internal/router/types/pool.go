package types

type Pool struct {
	Size    int   `json:"size"`
	Total   int64 `json:"total"`
	Running int64 `json:"running"`
	Waiting int64 `json:"waiting"`
}
