package snowflake

import (
	"fmt"
	"sync"
	"time"
)

const (
	Epoch           = 1730390400 // 元时间 2024-11-01 00:00:00
	timestampBits   = 31
	dataCenterBits  = 5
	nodeBits        = 7
	sequenceBits    = 10
	maxDataCenterID = (1 << dataCenterBits) - 1
	maxNodeID       = (1 << nodeBits) - 1
	maxSequence     = (1 << sequenceBits) - 1
	timestampShift  = sequenceBits + nodeBits + dataCenterBits
	dataCenterShift = sequenceBits + nodeBits
	nodeShift       = sequenceBits
	maxID           = (1 << (timestampBits + dataCenterBits + nodeBits)) - 1
)

// IDGenerator 共享内存和锁的结构
type IDGenerator struct {
	dataCenterID  int64
	nodeID        int64
	lastTimestamp int64
	sequence      int64
	lock          sync.Mutex
}

// 获取当前的秒级时间戳
func (g *IDGenerator) getTimestamp() int64 {
	return time.Now().Unix() - Epoch
}

// NextID 获取新的 ID
func (g *IDGenerator) NextID() int64 {
	g.lock.Lock()
	defer g.lock.Unlock()

	timestamp := g.getTimestamp()

	// 判断是否在同一秒内生成，若是则递增序列号
	if g.lastTimestamp == timestamp {
		g.sequence = (g.sequence + 1) & maxSequence
		// 如果序列号达到最大值，则等待下一秒
		if g.sequence == 0 {
			for timestamp <= g.lastTimestamp {
				timestamp = g.getTimestamp()
			}
		}
	} else {
		// 时间戳变更，序列号归零
		g.sequence = 0
	}

	g.lastTimestamp = timestamp

	// 通过移位运算拼接生成最终的 ID
	id := (timestamp << timestampShift) |
		(g.dataCenterID << dataCenterShift) |
		(g.nodeID << nodeShift) |
		g.sequence

	return id & maxID
}

// New 初始化雪花 ID 生成器
func New(dataCenterID, nodeID int64) (*IDGenerator, error) {
	if dataCenterID > maxDataCenterID || dataCenterID < 0 {
		return nil, fmt.Errorf("%d DataCenterID out of range", dataCenterID)
	}
	if nodeID > maxNodeID || nodeID < 0 {
		return nil, fmt.Errorf("%d NodeID out of range", nodeID)
	}
	return &IDGenerator{
		dataCenterID:  dataCenterID,
		nodeID:        nodeID,
		lastTimestamp: -1,
		sequence:      0,
	}, nil
}
