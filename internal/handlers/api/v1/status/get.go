package status

import (
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xmapst/osreapi/internal/handlers/base"
	"github.com/xmapst/osreapi/internal/handlers/types"
)

// Detail
// @Summary server state detail
// @description detail server state
// @Tags State
// @Accept json
// @Produce json
// @Success 200 {object} types.BaseRes
// @Failure 500 {object} types.BaseRes
// @Router /api/v1/state [get]
func Detail(c *gin.Context) {
	render := base.Gin{Context: c}
	render.SetJson(getStats())
}

var lastSampleTime time.Time
var lastPauseNs uint64 = 0
var lastNumGc uint32 = 0

var nsInMs = float64(time.Millisecond)

var statsMux sync.Mutex

func getStats() *types.Stats {
	statsMux.Lock()
	defer statsMux.Unlock()

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	now := time.Now()

	var gcPausePerSecond float64

	if lastPauseNs > 0 {
		pauseSinceLastSample := mem.PauseTotalNs - lastPauseNs
		gcPausePerSecond = float64(pauseSinceLastSample) / nsInMs
	}

	lastPauseNs = mem.PauseTotalNs

	countGc := int(mem.NumGC - lastNumGc)

	var gcPerSecond float64

	if lastNumGc > 0 {
		diff := float64(countGc)
		diffTime := now.Sub(lastSampleTime).Seconds()
		gcPerSecond = diff / diffTime
	}

	if countGc > 256 {
		// lagging GC pause times
		countGc = 256
	}

	gcPause := make([]float64, countGc)

	for i := 0; i < countGc; i++ {
		idx := int((mem.NumGC-uint32(i))+255) % 256
		pause := float64(mem.PauseNs[idx])
		gcPause[i] = pause / nsInMs
	}

	lastNumGc = mem.NumGC
	lastSampleTime = time.Now()

	return &types.Stats{
		Timestamp:    time.Now().UnixNano(),
		CpuNum:       runtime.NumCPU(),
		GoroutineNum: runtime.NumGoroutine(),
		Gomaxprocs:   runtime.GOMAXPROCS(0),
		CgoCallNum:   runtime.NumCgoCall(),
		// memory
		MemoryAlloc:      mem.Alloc,
		MemoryTotalAlloc: mem.TotalAlloc,
		MemorySys:        mem.Sys,
		MemoryLookups:    mem.Lookups,
		MemoryMallocs:    mem.Mallocs,
		MemoryFrees:      mem.Frees,
		// stack
		StackInUse: mem.StackInuse,
		// heap
		HeapAlloc:    mem.HeapAlloc,
		HeapSys:      mem.HeapSys,
		HeapIdle:     mem.HeapIdle,
		HeapInuse:    mem.HeapInuse,
		HeapReleased: mem.HeapReleased,
		HeapObjects:  mem.HeapObjects,
		// garbage collection
		GcNext:           mem.NextGC,
		GcLast:           mem.LastGC,
		GcNum:            mem.NumGC,
		GcPerSecond:      gcPerSecond,
		GcPausePerSecond: gcPausePerSecond,
		GcPause:          gcPause,
	}
}
