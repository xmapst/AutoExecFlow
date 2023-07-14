package types

type BaseRes struct {
	Code    int64       `json:"code" example:"255"`
	Message string      `json:"message,omitempty" example:"message"`
	Data    interface{} `json:"data,omitempty"`
}
