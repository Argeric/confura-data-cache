package main

import (
	"github.com/Conflux-Chain/confura-data-cache/cmd"
	"github.com/Conflux-Chain/go-conflux-util/config"
)

func main() {
	config.MustInit("confura_data_cache")
	cmd.Execute()
}
