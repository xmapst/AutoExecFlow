package types

type PoolSetting struct {
	Size int `json:"size" form:"command_type" binding:"required" example:"30"`
}
