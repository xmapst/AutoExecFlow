package types

type Healthyz struct {
	Server string `json:"server" yaml:"Server"`
	Client string `json:"client" yaml:"Client"`
	State  string `json:"state" yaml:"State"`
}
