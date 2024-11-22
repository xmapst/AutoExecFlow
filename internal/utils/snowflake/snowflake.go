package snowflake

import (
	"fmt"
	"sync"
	"time"
)

const (
	Epoch           = 1730390400 // 元时间 2024-11-01 00:00:00
	timestampBits   = 31         // 时间戳占用位数
	dataCenterBits  = 5          // 机房 ID 占用位数
	nodeBits        = 7          // 节点 ID 占用位数
	sequenceBits    = 10         // 递增序列占用位数
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
	node      uint64
	timestamp uint64
	sequence  uint64
	lock      sync.Mutex
}

// NextID 获取新的 ID
func (g *IDGenerator) NextID() int64 {
	g.lock.Lock()
	defer g.lock.Unlock()

	// 检查序列是否溢出
	if g.sequence > maxSequence {
		// 等待时间前进
		for g.timestamp > uint64(time.Now().Unix())-Epoch {
			time.Sleep(100 * time.Millisecond)
		}
		// 时间递增
		g.timestamp++
		// 序列重置
		g.sequence = 0
	} else {
		g.sequence++
	}

	// 通过移位运算拼接生成最终的 ID
	id := (g.timestamp << timestampShift) | g.node | g.sequence

	return int64(id & maxID)
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
		node:      uint64((dataCenterID << dataCenterShift) | (nodeID << nodeShift)),
		timestamp: uint64(time.Now().Unix() - Epoch),
		sequence:  0,
	}, nil
}
