package main

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/adshao/go-binance/v2"

	"binance-bot-test/config"
)

func replaceBid(newBid *binance.Bid, bids map[string]string) {

	bids[newBid.Price] = newBid.Quantity
}

func replaceAsk(newAsk *binance.Ask, asks map[string]string) {

	asks[newAsk.Price] = newAsk.Quantity
}

func displayOrderBook(bids map[string]string, asks map[string]string) {

	fmt.Print("\033[H\033[2J")

	bidsSorted := make([]string, 0)
	for k, _ := range bids {
		bidsSorted = append(bidsSorted, k)
	}
	sort.Strings(bidsSorted)

	asksSorted := make([]string, 0)
	for k, _ := range asks {
		asksSorted = append(asksSorted, k)
	}
	sort.Strings(asksSorted)

	b := len(bidsSorted) - 1
	for row := 0; row < 30; row++ {
		fmt.Printf("%s (%s)\t %s (%s)\n", bidsSorted[b], bids[bidsSorted[b]], asksSorted[row], asks[asksSorted[row]])
		b--
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
				replaceBid(&bid, bids)
			}
			for _, ask := range event.Asks {
				replaceAsk(&ask, asks)
			}
			displayOrderBook(bids, asks)
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
