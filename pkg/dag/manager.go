package dag

import (
	"fmt"
	"sync"
)

var manager = sync.Map{}

const (
	graphPrefix  = "graph#%s#graph"
	vertexPrefix = "graph#%s#vertex#%s#vertex#graph"
)

type IControl interface {
	Name() string
	Kill() error
	Pause(duration string) error
	Resume()
	State() State
	WaitResume()
}

func leave(key string) (IControl, error) {
	value, ok := manager.Load(key)
	if !ok {
		return nil, ErrNotFound
	}
	m, ok := value.(IControl)
	if !ok {
		manager.Delete(key)
		return nil, ErrWrongType
	}
	return m, nil
}

func join(key string, iManager IControl) {
	manager.Store(key, iManager)
}

func remove(key string) {
	manager.Delete(key)
}

func GraphManager(gName string) (IControl, error) {
	return leave(fmt.Sprintf(graphPrefix, gName))
}

func VertexManager(gName, vName string) (IControl, error) {
	return leave(fmt.Sprintf(vertexPrefix, gName, vName))
}
