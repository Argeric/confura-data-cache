package cmd

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Conflux-Chain/go-conflux-util/viper"
	"github.com/openweb3/web3go"
	"github.com/sirupsen/logrus"
)

// DataContext context to hold sdk clients for blockchain interoperation.
type DataContext struct {
	Eth *web3go.Client
}

type SdkConfig struct {
	URL             string
	Retry           int
	RetryInterval   time.Duration `default:"1s"`
	RequestTimeout  time.Duration `default:"3s"`
	MaxConnsPerHost int           `default:"1024"`
}

func MustInitDataContext() DataContext {
	sdkCfg := SdkConfig{}
	viper.MustUnmarshalKey("eth", &sdkCfg)
	opt := web3go.ClientOption{}
	opt.WithRetry(sdkCfg.Retry, sdkCfg.RetryInterval).
		WithTimout(sdkCfg.RequestTimeout).
		WithMaxConnectionPerHost(sdkCfg.MaxConnsPerHost)
	eth := web3go.MustNewClientWithOption(sdkCfg.URL, opt)

	return DataContext{
		Eth: eth,
	}
}

func (ctx *DataContext) Close() {
	if ctx.Eth != nil {
		ctx.Eth.Close()
	}
}

func GracefulShutdown(wg *sync.WaitGroup, cancel context.CancelFunc) {
	// Handle sigterm and await termChan signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGTERM, syscall.SIGINT)

	// Wait for SIGTERM to be captured
	<-termChan
	logrus.Info("SIGTERM/SIGINT received, shutdown process initiated")

	// Cancel to notify active goroutines to clean up.
	cancel()

	logrus.Info("Waiting for shutdown...")
	wg.Wait()

	logrus.Info("Shutdown gracefully")
}
