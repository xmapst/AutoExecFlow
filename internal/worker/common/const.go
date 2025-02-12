package common

import "strings"

const ConsoleStart = "OSREAPI::CONSOLE::START"
const ConsoleDone = "OSREAPI::CONSOLE::DONE"

type Action int

const (
	ActionUnknown Action = iota
	// ActionAllow 允许执行
	ActionAllow
	// ActionBlock 禁止执行
	ActionBlock
	// ActionSkip 跳过执行
	ActionSkip
)

func (a Action) String() string {
	switch a {
	case ActionUnknown:
		return "unknown"
	case ActionAllow:
		return "allow"
	case ActionBlock:
		return "block"
	case ActionSkip:
		return "skip"
	default:
		return "unknown"
	}
}

func ActionConvert(action string) Action {
	action = strings.ToLower(action)
	switch action {
	case "allow":
		return ActionAllow
	case "block":
		return ActionBlock
	case "skip":
		return ActionSkip
	default:
		return ActionUnknown
	}
}
