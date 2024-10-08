package service

import (
	"fmt"
	"slices"
	"strings"

	"github.com/xmapst/AutoExecFlow/internal/storage/models"
	"github.com/xmapst/AutoExecFlow/types"
)

func ConvertState(state models.State) types.Code {
	switch state {
	case models.StateStop:
		return types.CodeSuccess
	case models.StateRunning:
		return types.CodeRunning
	case models.StatePending:
		return types.CodePending
	case models.StatePaused:
		return types.CodePaused
	case models.StateFailed:
		return types.CodeFailed
	default:
		return types.CodeNoData
	}
}

func GenerateStateMessage(baseMessage string, groups map[models.State][]string) string {
	var keys []models.State
	for k := range groups {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	var messages []string
	if baseMessage != "" {
		messages = append(messages, baseMessage)
	}
	for _, key := range keys {
		messages = append(messages, fmt.Sprintf("%s: [%s]", models.StateMap[key], strings.Join(groups[key], ",")))
	}
	return strings.Join(messages, "; ")
}
