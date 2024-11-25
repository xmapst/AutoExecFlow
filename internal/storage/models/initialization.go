package models

import (
	"sync"

	"github.com/xmapst/AutoExecFlow/internal/utils/snowflake"
)

var (
	NodeID       int64
	DataCenterID int64
	idGenerator  sync.Map
)

func getNextID(name string) (int64, error) {
	gen, ok := idGenerator.Load(name)
	if ok {
		return gen.(*snowflake.IDGenerator).NextID().Int64(), nil
	}
	var err error
	gen, err = snowflake.New(DataCenterID, NodeID)
	if err != nil {
		return 0, err
	}
	idGenerator.Store(name, gen)
	return gen.(*snowflake.IDGenerator).NextID().Int64(), nil
}
