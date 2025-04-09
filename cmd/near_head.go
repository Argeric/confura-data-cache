package cmd

import (
	"context"
	"sync"

	cdcSync "github.com/Conflux-Chain/confura-data-cache/nearhead/sync"
	"github.com/spf13/cobra"
)

var (
	syncCmd = &cobra.Command{
		Use:   "nearhead",
		Short: "Start cache near head data",
		Run:   startCacheService,
	}
)

func init() {
	rootCmd.AddCommand(syncCmd)
}

func startCacheService(*cobra.Command, []string) {
	dataCtx := MustInitDataContext()
	defer dataCtx.Close()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	syncers := make([]*cdcSync.EthSyncer, 0)
	for _, eth := range dataCtx.Eths {
		syncer := cdcSync.MustNewEthSyncer(eth)
		syncers = append(syncers, syncer)
		go syncer.Start(ctx, &wg)
	}

	GracefulShutdown(&wg, cancel)
}
