package orderbook

import (
	"fmt"
	"sort"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/common"
)

type FloatPriceLevel struct {
	PriceLevel common.PriceLevel
	Price      float64
	Quantity   float64
}

var (
	bids = map[string]FloatPriceLevel{}
	asks = map[string]FloatPriceLevel{}
)

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

func Initialise(depthSnapshot *binance.DepthResponse) {
	for _, bid := range depthSnapshot.Bids {

		price, quantity, err := bid.Parse()
		if err != nil {
			fmt.Printf("Error parsing bid: %s\n", err)
		} else {
			bids[bid.Price] = FloatPriceLevel{bid, price, quantity}
		}
	}

	for _, ask := range depthSnapshot.Asks {

		price, quantity, err := ask.Parse()
		if err != nil {
			fmt.Printf("Error parsing ask: %s\n", err)
		} else {
			asks[ask.Price] = FloatPriceLevel{ask, price, quantity}
		}
	}
}

func Display(bids map[string]FloatPriceLevel, asks map[string]FloatPriceLevel) {

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

func Update(depthSnapshot *binance.DepthResponse, symbol string) {
	lastUpdateID := depthSnapshot.LastUpdateID
	wsDepthHandler := func(event *binance.WsDepthEvent) {

		if (event.LastUpdateID > lastUpdateID) && (event.FirstUpdateID == (lastUpdateID + 1)) {

			for _, bid := range event.Bids {
				replaceBid(&bid, bids)
			}
			for _, ask := range event.Asks {
				replaceAsk(&ask, asks)
			}
			Display(bids, asks)
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
