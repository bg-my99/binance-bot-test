package main

import (
	"context"
	"fmt"

	"github.com/adshao/go-binance/v2"

	"binance-bot-test/config"
	"binance-bot-test/order-book"
)

func main() {

	var cfg config.Config
	config.ReadEnv(&cfg)
	fmt.Printf("%+v\n", cfg)

	whichNet := "main network"
	if cfg.UseTestnet {
		whichNet = "testnet"
	}
	fmt.Printf("Connecting to %s\n", whichNet)
	binance.UseTestnet = cfg.UseTestnet

	client := binance.NewClient(cfg.AccessKeys.ApiKey, cfg.AccessKeys.SecretKey)
	symbol := "LTCBTC"
	depthSnapshot, err := client.NewDepthService().Symbol(symbol).Limit(1000).Do(context.Background())

	if err != nil {
		fmt.Printf("Error requesting depthSnapshot: %s\n", err)
	} else {
		orderbook.Initialise(depthSnapshot)
		orderbook.Update(depthSnapshot, symbol)
	}
}
