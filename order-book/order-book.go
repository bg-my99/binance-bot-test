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

type OrderBook struct {
	symbol string
	bids map[string]FloatPriceLevel
	asks map[string]FloatPriceLevel

	depthSnapshot *binance.DepthResponse
}

func (o *OrderBook) replaceBid(newBid *binance.Bid) {

	price, quantity, err := newBid.Parse()
	if err != nil {
		fmt.Printf("Error parsing bid: %s\n", err)
	} else {

		if quantity == 0.0 {
			delete(o.bids, newBid.Price)
		} else {
			o.bids[newBid.Price] = FloatPriceLevel{*newBid, price, quantity}
		}
	}
}

func (o *OrderBook) replaceAsk(newAsk *binance.Ask) {

	price, quantity, err := newAsk.Parse()
	if err != nil {
		fmt.Printf("Error parsing bid: %s\n", err)
	} else {

		if quantity == 0.0 {
			delete(o.asks, newAsk.Price)
		} else {
			o.asks[newAsk.Price] = FloatPriceLevel{*newAsk, price, quantity}
		}
	}
}

func (o *OrderBook) Initialise(depthSnapshot *binance.DepthResponse, symbol string) {
	o.symbol = symbol
	o.bids = map[string]FloatPriceLevel{}
	for _, bid := range depthSnapshot.Bids {

		price, quantity, err := bid.Parse()
		if err != nil {
			fmt.Printf("Error parsing bid: %s\n", err)
		} else {
			o.bids[bid.Price] = FloatPriceLevel{bid, price, quantity}
		}
	}

	o.asks = map[string]FloatPriceLevel{}
	for _, ask := range depthSnapshot.Asks {

		price, quantity, err := ask.Parse()
		if err != nil {
			fmt.Printf("Error parsing ask: %s\n", err)
		} else {
			o.asks[ask.Price] = FloatPriceLevel{ask, price, quantity}
		}
	}
	o.depthSnapshot = depthSnapshot
}

func (o *OrderBook) Display() {

	fmt.Print("\033[H\033[2J")

	bidsSorted := make([]string, 0)
	for k, _ := range o.bids {
		bidsSorted = append(bidsSorted, k)
	}
	sort.Strings(bidsSorted)

	asksSorted := make([]string, 0)
	for k, _ := range o.asks {
		asksSorted = append(asksSorted, k)
	}
	sort.Strings(asksSorted)

	b := len(bidsSorted) - 1
	for row := 0; row < 30; row++ {
		fmt.Printf("%s (%v)\t %s (%v)\n", bidsSorted[b], o.bids[bidsSorted[b]], asksSorted[row], o.asks[asksSorted[row]])
		b--
	}
}

func (o *OrderBook) Update() {
	lastUpdateID := o.depthSnapshot.LastUpdateID
	wsDepthHandler := func(event *binance.WsDepthEvent) {

		if (event.LastUpdateID > lastUpdateID) && (event.FirstUpdateID == (lastUpdateID + 1)) {

			for _, bid := range event.Bids {
				o.replaceBid(&bid)
			}
			for _, ask := range event.Asks {
				o.replaceAsk(&ask)
			}
			o.Display()
		}
		lastUpdateID = event.LastUpdateID
	}
	errHandler := func(err error) {
		fmt.Println(err)
	}
	doneC, stopC, err := binance.WsDepthServe(o.symbol, wsDepthHandler, errHandler)
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
