package sync

import (
	"context"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sync"
	"time"

	"github.com/Conflux-Chain/confura-data-cache/nearhead/cache"
	"github.com/openweb3/web3go"
	"github.com/openweb3/web3go/types"
)

var (
	intervalNormal    = time.Second
	intervalException = time.Second * 3
)

type EthSyncer struct {
	eth          *web3go.Client
	memory       *cache.EthCache
	currentBlock uint64
}

func MustNewEthSyncer(eth *web3go.Client) *EthSyncer {
	if eth == nil {
		logrus.Fatal("No sdk client provided!")
	}

	memory := cache.MustNewEthCache()

	return &EthSyncer{
		eth:    eth,
		memory: memory,
	}
}

func (s *EthSyncer) Start(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	ticker := time.NewTicker(intervalNormal)
	defer ticker.Stop()

	logrus.Info("Near head syncer starting to cache data")
	for {
		select {
		case <-ctx.Done():
			logrus.Info("Near head syncer shutdown ok")
			return
		case <-ticker.C:
			if err := s.cache(ctx, ticker); err != nil {
				logrus.WithError(err).WithField("currentBlock", s.currentBlock).Warn("Failed to cache data")
			}
		}
	}
}

func (s *EthSyncer) cache(ctx context.Context, ticker *time.Ticker) error {
	if s.currentBlock == 0 {
		finalized, err := GetBlockByNumber(s.eth, types.FinalizedBlockNumber, false)
		if err != nil {
			return err
		}
		s.currentBlock = finalized.Number.Uint64() + 1
	}

	latest, err := s.eth.Eth.BlockNumber()
	if err != nil {
		return err
	}

	for bn := s.currentBlock; bn <= latest.Uint64(); bn++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := s.cacheBlock(bn); err != nil {
				ticker.Reset(intervalException)
				return err
			}
		}
	}

	return nil
}

func (s *EthSyncer) cacheBlock(blockNumber uint64) error {
	block, err := GetBlockByNumber(s.eth, types.BlockNumber(blockNumber), true)
	if err != nil {
		return err
	}
	s.memory.AddBlock(block)

	receipts, err := GetBlockReceipts(s.eth, types.BlockNumber(blockNumber))
	if err != nil {
		return err
	}
	s.memory.AddReceipts(blockNumber, receipts)

	traces, err := GetBlockTraces(s.eth, types.BlockNumber(blockNumber))
	if err != nil {
		return err
	}
	s.memory.AddTraces(blockNumber, traces)

	return nil
}

func GetBlockByNumber(w3c *web3go.Client, blockNumber types.BlockNumber, isFull bool) (*types.Block, error) {
	block, err := w3c.Eth.BlockByNumber(blockNumber, isFull)
	if err != nil {
		return nil, errors.WithMessagef(err, "Failed to get block %v", blockNumber)
	}

	if block == nil {
		return nil, errors.Errorf("Invalid nil block %v", blockNumber)
	}

	return block, nil
}

func GetBlockReceipts(w3c *web3go.Client, blockNumber types.BlockNumber) ([]types.Receipt, error) {
	blockNumOrHash := types.BlockNumberOrHashWithNumber(blockNumber)
	receipts, err := w3c.Parity.BlockReceipts(&blockNumOrHash)
	if err != nil {
		return nil, errors.WithMessagef(err, "Failed to get block receipts %v", blockNumber)
	}

	if receipts == nil {
		return nil, errors.Errorf("Invalid nil block receipts %v", blockNumber)
	}

	return receipts, nil
}

func GetBlockTraces(w3c *web3go.Client, blockNumber types.BlockNumber) ([]types.LocalizedTrace, error) {
	blockNumOrHash := types.BlockNumberOrHashWithNumber(blockNumber)
	traces, err := w3c.Trace.Blocks(blockNumOrHash)
	if err != nil {
		return nil, errors.WithMessagef(err, "Failed to get block receipts %v", blockNumber)
	}

	if traces == nil {
		return nil, errors.Errorf("Invalid nil block receipts %v", blockNumber)
	}

	return traces, nil
}
