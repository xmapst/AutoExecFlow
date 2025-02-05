package dag

import (
	"errors"
)

var (
	ErrContext   = errors.New("context is null")
	ErrForceKill = errors.New("was forcibly terminated")
	ErrNotFound  = errors.New("not found or closed")
	ErrRunning   = errors.New("is running, can't pause")
	ErrWrongType = errors.New("wrong interface type, cannot be closed")

	ErrCycleDetected       = errors.New("dependency cycle detected")
	ErrDuplicateVertexName = errors.New("duplicate vertex name")
	ErrDuplicateCompile    = errors.New("duplicate compile graph")
	ErrEmptyGraph          = errors.New("the vertex of the graph is null")
)
