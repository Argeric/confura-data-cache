package cache

import (
	"container/list"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"sync"

	"github.com/Conflux-Chain/confura-data-cache/nearhead/config"
	viperUtil "github.com/Conflux-Chain/go-conflux-util/viper"
	"github.com/openweb3/web3go/types"
)

type EthCache struct {
	config   *config.Config
	blocks   *BlocksCache
	receipts *ReceiptsCache
	traces   *TracesCache
}

type BlocksCache struct {
	cache map[uint64]*types.Block
	bns   *list.List
	size  uint64
	mutex sync.Mutex
}

type ReceiptsCache struct {
	cache map[uint64][]types.Receipt
	bns   *list.List
	size  uint64
	mutex sync.Mutex
}

type TracesCache struct {
	cache map[uint64][]types.LocalizedTrace
	bns   *list.List
	size  uint64
	mutex sync.Mutex
}

func MustNewEthCache() *EthCache {
	var cfg config.Config
	viperUtil.MustUnmarshalKey("nearHead", &cfg)

	blocks := BlocksCache{
		cache: make(map[uint64]*types.Block),
		bns:   list.New(),
	}
	receipts := ReceiptsCache{
		cache: make(map[uint64][]types.Receipt),
		bns:   list.New(),
	}
	traces := TracesCache{
		cache: make(map[uint64][]types.LocalizedTrace),
		bns:   list.New(),
	}

	return &EthCache{
		config:   &cfg,
		blocks:   &blocks,
		receipts: &receipts,
		traces:   &traces,
	}
}

func (c *EthCache) AddBlock(block *types.Block) {
	c.blocks.mutex.Lock()
	defer c.blocks.mutex.Unlock()

	blockj, _ := json.Marshal(block)
	logrus.WithFields(logrus.Fields{
		"bn":    block.Number.Uint64(),
		"block": string(blockj),
	}).Info("debug add block ===1===")

	blockSize := block.Size
	c.evictIfNeeded(c.blocks.size, blockSize, c.config.CacheSizeBlocks, c.blocks, c.blocks.bns)

	c.blocks.bns.PushFront(block.Number)
	c.blocks.cache[block.Number.Uint64()] = block
	c.blocks.size += blockSize
}

func (c *EthCache) AddReceipts(blockNumber uint64, receipts []types.Receipt) {
	c.receipts.mutex.Lock()
	defer c.receipts.mutex.Unlock()

	receiptsj, _ := json.Marshal(receipts)
	logrus.WithFields(logrus.Fields{
		"bn":       blockNumber,
		"receipts": string(receiptsj),
	}).Info("debug add receipts ===2===")

	receiptsSize := uint64(0)
	c.evictIfNeeded(c.receipts.size, receiptsSize, c.config.CacheSizeReceipts, c.receipts, c.receipts.bns)

	c.receipts.bns.PushFront(blockNumber)
	c.receipts.cache[blockNumber] = receipts
	c.receipts.size += receiptsSize
}

func (c *EthCache) AddTraces(blockNumber uint64, traces []types.LocalizedTrace) {
	c.traces.mutex.Lock()
	defer c.traces.mutex.Unlock()

	tracesj, _ := json.Marshal(traces)
	logrus.WithFields(logrus.Fields{
		"bn":     blockNumber,
		"traces": string(tracesj),
	}).Info("debug add traces ===3===")

	traceSize := uint64(0)
	c.evictIfNeeded(c.traces.size, traceSize, c.config.CacheSizeTraces, c.traces.cache, c.traces.bns)

	c.traces.bns.PushFront(blockNumber)
	c.traces.cache[blockNumber] = traces
	c.traces.size += traceSize
}

func (c *EthCache) evictIfNeeded(currentSize uint64, newItemSize uint64, maxSize uint64,
	cache interface{}, l *list.List) {
	if currentSize+newItemSize > maxSize {
		for e := l.Back(); e != nil; e = e.Prev() {
			blockNumber := e.Value.(uint64)
			switch c := cache.(type) {
			case map[uint64]*types.Block:
				delete(c, blockNumber)
			case map[uint64][]types.Receipt:
				delete(c, blockNumber)
			case map[uint64][]types.LocalizedTrace:
				delete(c, blockNumber)
			}

			l.Remove(e)

			if currentSize+newItemSize <= maxSize {
				break
			}
		}
	}
}
