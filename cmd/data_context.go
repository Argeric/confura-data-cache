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
	Eths []*web3go.Client
}

type ClientsConfig struct {
	Http            []string
	Retry           int
	RetryInterval   time.Duration `default:"1s"`
	RequestTimeout  time.Duration `default:"3s"`
	MaxConnsPerHost int           `default:"1024"`
}

func MustInitDataContext() DataContext {
	clientsConfig := ClientsConfig{}
	viper.MustUnmarshalKey("eth", &clientsConfig)

	opt := web3go.ClientOption{}
	opt.WithRetry(clientsConfig.Retry, clientsConfig.RetryInterval).
		WithTimout(clientsConfig.RequestTimeout).
		WithMaxConnectionPerHost(clientsConfig.MaxConnsPerHost)

	eths := make([]*web3go.Client, 0)
	for _, http := range clientsConfig.Http {
		eth := web3go.MustNewClientWithOption(http, opt)
		eths = append(eths, eth)
	}

	return DataContext{
		Eths: eths,
	}
}

func (ctx *DataContext) Close() {
	for _, eth := range ctx.Eths {
		if eth != nil {
			eth.Close()
		}
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
