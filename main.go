package main

import (
	"context"
	"fmt"
	"sort"
	//"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/common"

	"binance-bot-test/config"
)

type FloatPriceLevel struct {
	PriceLevel common.PriceLevel
	Price      float64
	Quantity   float64
}

func replaceBid(newBid *binance.Bid, bids map[string]FloatPriceLevel) {

	price, quantity, err := newBid.Parse()
	if err != nil {
		fmt.Printf("Error parsing bid: %s\n", err)
	} else {

		if quantity == 0.0 {
			delete(bids, newBid.Price)
		} else {
			bids[newBid.Price] = FloatPriceLevel{*newBid, price, quantity}
		}
	}
}

func replaceAsk(newAsk *binance.Ask, asks map[string]FloatPriceLevel) {

	price, quantity, err := newAsk.Parse()
	if err != nil {
		fmt.Printf("Error parsing bid: %s\n", err)
	} else {

		if quantity == 0.0 {
			delete(asks, newAsk.Price)
		} else {
			asks[newAsk.Price] = FloatPriceLevel{*newAsk, price, quantity}
		}
	}
}

func displayOrderBook(bids map[string]FloatPriceLevel, asks map[string]FloatPriceLevel) {

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

	bids := map[string]FloatPriceLevel{}
	for _, bid := range depthSnapshot.Bids {

		price, quantity, err := bid.Parse()
		if err != nil {
			fmt.Printf("Error parsing bid: %s\n", err)
		} else {
			bids[bid.Price] = FloatPriceLevel{bid, price, quantity}
		}
	}

	asks := map[string]FloatPriceLevel{}
	for _, ask := range depthSnapshot.Asks {

		price, quantity, err := ask.Parse()
		if err != nil {
			fmt.Printf("Error parsing ask: %s\n", err)
		} else {
			asks[ask.Price] = FloatPriceLevel{ask, price, quantity}
		}
	}

	lastUpdateID := depthSnapshot.LastUpdateID
	wsDepthHandler := func(event *binance.WsDepthEvent) {

		if (event.LastUpdateID > lastUpdateID) && (event.FirstUpdateID == (lastUpdateID + 1)) {

			for _, bid := range event.Bids {
				replaceBid(&bid, bids)
			}
			for _, ask := range event.Asks {
				replaceAsk(&ask, asks)
			}
			displayOrderBook(bids, asks)
		}
		lastUpdateID = event.LastUpdateID
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
		time.Sleep(50 * time.Second)
		stopC <- struct{}{}
	}()
	// remove this if you do not want to be blocked here
	<-doneC
}
