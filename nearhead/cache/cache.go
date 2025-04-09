package cache

import (
	"container/list"
	"encoding/json"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/openweb3/web3go/types"
	"github.com/sirupsen/logrus"
)

type EthCache struct {
	blocks   map[uint64]*types.Block
	receipts map[uint64][]*types.Receipt
	traces   map[uint64][]types.LocalizedTrace

	blockNumbers list.List
	blockHashes  map[common.Hash]uint64
	transactions map[common.Hash]Transaction

	mutex sync.Mutex
}

func MustNewEthCache() *EthCache {
	return &EthCache{
		blockNumbers: *list.New(),
		blocks:       make(map[uint64]*types.Block),
		transactions: make(map[common.Hash]Transaction),
		receipts:     make(map[uint64][]*types.Receipt),
		traces:       make(map[uint64][]types.LocalizedTrace),
	}
}

func (c *EthCache) Set(block *types.Block, receipts []*types.Receipt, traces []types.LocalizedTrace) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	bn := block.Number.Uint64()
	c.blocks[bn] = block
	c.receipts[bn] = receipts
	c.traces[bn] = traces

	c.blockNumbers.PushFront(bn)
	c.blockHashes[block.Hash] = bn
	for _, tx := range block.Transactions.Transactions() {
		c.transactions[tx.Hash] = Transaction{
			blockNumber:      bn,
			transactionIndex: *tx.TransactionIndex,
		}
	}

	blockj, _ := json.Marshal(block)
	receiptsj, _ := json.Marshal(receipts)
	tracesj, _ := json.Marshal(traces)
	logrus.WithFields(logrus.Fields{
		"bn":       block.Number.Uint64(),
		"block":    string(blockj),
		"receipts": string(receiptsj),
		"traces":   string(tracesj),
	}).Info("debug set cache ===1===")

	return nil
}

func (c *EthCache) GetBlockByNumber(blockNumber uint64, isFull bool) (*types.Block, bool) {
	block, exists := c.blocks[blockNumber]
	if !exists {
		return nil, false
	}

	if !isFull {
		txOrHashList := types.NewTxOrHashListByHashes(block.Transactions.Hashes())
		block.Transactions = *txOrHashList
	}

	return block, exists
}

func (c *EthCache) GetBlockByHash(blockHash common.Hash, isFull bool) (*types.Block, bool) {
	blockNumber, exists := c.blockHashes[blockHash]
	if !exists {
		return nil, false
	}

	block, exists := c.blocks[blockNumber]
	if !exists {
		return nil, false
	}

	if !isFull {
		txOrHashList := types.NewTxOrHashListByHashes(block.Transactions.Hashes())
		block.Transactions = *txOrHashList
	}

	return block, exists
}

func (c *EthCache) GetTransactionByHash(txHash common.Hash) (*types.TransactionDetail, bool) {
	txCache, exists := c.transactions[txHash]
	if !exists {
		return nil, false
	}

	block, exists := c.blocks[txCache.blockNumber]
	if !exists {
		return nil, false
	}

	tx := block.Transactions.Transactions()[txCache.transactionIndex]
	return &tx, true
}

func (c *EthCache) GetBlockReceipts(blockNumber uint64) ([]*types.Receipt, bool) {
	receipts, exists := c.receipts[blockNumber]
	return receipts, exists
}

func (c *EthCache) GetTransactionReceipt(txHash common.Hash) (*types.Receipt, bool) {
	txCache, exists := c.transactions[txHash]
	if !exists {
		return nil, false
	}

	receipts, exists := c.receipts[txCache.blockNumber]
	if !exists {
		return nil, false
	}

	receipt := receipts[txCache.transactionIndex]
	return receipt, true
}

func (c *EthCache) GetBlockTraces(blockNumber uint64) ([]types.LocalizedTrace, bool) {
	traces, exists := c.traces[blockNumber]
	return traces, exists
}

func (c *EthCache) GetTransactionTraces(txHash common.Hash) ([]types.LocalizedTrace, bool) {
	txCache, exists := c.transactions[txHash]
	if !exists {
		return nil, false
	}

	traces, exists := c.traces[txCache.blockNumber]

	txTraces := make([]types.LocalizedTrace, 0)
	for _, trace := range traces {
		if txHash == *trace.TransactionHash {
			txTraces = append(txTraces, trace)
		}
	}

	return txTraces, exists
}

type Transaction struct {
	blockNumber      uint64
	transactionIndex uint64
}
