package cache

import (
	"container/list"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/openweb3/web3go/types"
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
		blocks:   make(map[uint64]*types.Block),
		receipts: make(map[uint64][]*types.Receipt),
		traces:   make(map[uint64][]types.LocalizedTrace),

		blockNumbers: *list.New(),
		blockHashes:  make(map[common.Hash]uint64),
		transactions: make(map[common.Hash]Transaction),
	}
}

func (c *EthCache) Set(block *types.Block, receipts []*types.Receipt, traces []types.LocalizedTrace) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// TODO evict if exceeds maxsize

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

	return nil
}

func (c *EthCache) GetBlockByNumber(blockNumber uint64, isFull bool) (*types.Block, bool) {
	block, exists := c.blocks[blockNumber]
	if !exists {
		return nil, false
	}

	if !isFull {
		hashes := make([]common.Hash, 0)
		for _, tx := range block.Transactions.Transactions() {
			hashes = append(hashes, tx.Hash)
		}
		txOrHashList := types.NewTxOrHashListByHashes(hashes)

		blockClone := *block
		blockClone.Transactions = *txOrHashList
		return &blockClone, exists
	}

	return block, exists
}

func (c *EthCache) GetBlockByHash(blockHash common.Hash, isFull bool) (*types.Block, bool) {
	blockNumber, exists := c.blockHashes[blockHash]
	if !exists {
		return nil, false
	}

	return c.GetBlockByNumber(blockNumber, isFull)
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
