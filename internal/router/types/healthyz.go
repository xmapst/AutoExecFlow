package types

type Healthyz struct {
	Server string `json:"server"`
	Client string `json:"client"`
	State  string `json:"state"`
}
