// Package drivers multiple storage engines
package drivers

import (
	"github.com/xmapst/AutoExecFlow/pkg/glua-libs/storage/drivers/interfaces"
	memory "github.com/xmapst/AutoExecFlow/pkg/glua-libs/storage/drivers/memory"
)

var (
	knownDrivers = make(map[string]interfaces.Driver, 0)
)

func init() {
	knownDrivers[`memory`] = &memory.Storage{}
}

func Get(name string) (interfaces.Driver, bool) {
	// read only struct
	d, ok := knownDrivers[name]
	return d, ok
}
