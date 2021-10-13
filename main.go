package main

import (
	"context"
	"fmt"
	"time"

	"github.com/adshao/go-binance/v2"

	"binance-bot-test/config"
)

func replaceBid(newBid *binance.Bid, depth *binance.DepthResponse) {

	for i, bid := range depth.Bids {
		if newBid.Price == bid.Price {
			depth.Bids[i] = *newBid
		}
	}
}

func replaceAsk(newAsk *binance.Ask, depth *binance.DepthResponse) {

	for i, ask := range depth.Asks {
		if newAsk.Price == ask.Price {
			depth.Asks[i] = *newAsk
		}
	}
}

func displayOrderBook(depth *binance.DepthResponse) {

	fmt.Print("\033[H\033[2J")

	for row := 0; row < 30; row++ {
		fmt.Printf("%v\t %v\n", depth.Bids[row], depth.Asks[row])
	}
}

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

	bids := map[string]string{}
	for _, bid := range depthSnapshot.Bids {
		bids[bid.Price] = bid.Quantity
	}

	asks := map[string]string{}
	for _, ask := range depthSnapshot.Asks {
		asks[ask.Price] = ask.Quantity
	}

	wsDepthHandler := func(event *binance.WsDepthEvent) {

		if event.LastUpdateID > depthSnapshot.LastUpdateID {

			for _, bid := range event.Bids {
				replaceBid(&bid, depthSnapshot)
			}
			for _, ask := range event.Asks {
				replaceAsk(&ask, depthSnapshot)
			}
			displayOrderBook(depthSnapshot)
		}
	}
	errHandler := func(err error) {
		fmt.Println(err)
	}
	doneC, stopC, err := binance.WsDepthServe(symbol, wsDepthHandler, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		time.Sleep(5 * time.Second)
		stopC <- struct{}{}
	}()
	// remove this if you do not want to be blocked here
	<-doneC
}
