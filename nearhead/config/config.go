package config

type Config struct {
	CacheSizeBlocks   uint64 `default:"100"` // in MB
	CacheSizeTxs      uint64 `default:"100"` // in MB
	CacheSizeReceipts uint64 `default:"100"` // in MB
	CacheSizeTraces   uint64 `default:"100"` // in MB
}
