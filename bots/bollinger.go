package bots

import (
	"fmt"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"

	"binance-bot-test/calcs"
	candles "binance-bot-test/storage"
)

const numPoints = 20

type BollingerBot struct {
	PnL     float64
	inTrade bool

	Trades  []Trade
	candles candles.Candles
}

func (b *BollingerBot) Init() {

	b.candles = candles.Candles{}
	b.candles.Init(int64((time.Second * 300) / time.Millisecond))
	b.inTrade = false
}

func (b *BollingerBot) AddMarketTrade(trade *binance.AggTrade) {
	b.candles.AddTrade(trade)

	pts, _ := b.candles.GetSortedCandles()
	movingAverage := calcs.GetMovingAverage(pts, numPoints)
	standardDeviation := calcs.GetStandardDeviation(pts, movingAverage, numPoints)

	if len(movingAverage) > 0 {
		price, _ := strconv.ParseFloat(trade.Price, 64)

		if price > (movingAverage[movingAverage.Len()-1].Y + (2.0 * standardDeviation[standardDeviation.Len()-1].Y)) {
			if b.inTrade && (len(b.Trades) > 0) {
				fmt.Println("**Price is above line")
				b.inTrade = false
				b.Trades = append(b.Trades, Trade{Price: price, Timestamp: trade.Timestamp, Type: 2})
				fmt.Printf("Sell at %s\n", trade.Price)
				//fmt.Printf("MA: len %d, p:%f\n", len(movingAverage), movingAverage[movingAverage.Len()-1].Y)
				//fmt.Printf("SD: len %d, p:%f\n", len(standardDeviation), standardDeviation[standardDeviation.Len()-1].Y)
			}
		} else if price < (movingAverage[movingAverage.Len()-1].Y - (2.0 * standardDeviation[standardDeviation.Len()-1].Y)) {
			if !b.inTrade {
				fmt.Println("**Price is below line")
				b.inTrade = true
				b.Trades = append(b.Trades, Trade{Price: price, Timestamp: trade.Timestamp, Type: 1})
				fmt.Printf("Buy at %s\n", trade.Price)
				//fmt.Printf("MA: len %d, p:%f\n", len(movingAverage), movingAverage[movingAverage.Len()-1].Y)
				//fmt.Printf("SD: len %d, p:%f\n", len(standardDeviation), standardDeviation[standardDeviation.Len()-1].Y)
			}
		}
	}
}

func (b *BollingerBot) DisplayPnL(tradeQuantity float64) {
	b.PnL = tradeQuantity
	amountHeld := 0.0
	for _, trade := range b.Trades {
		if trade.Type == 1 {
			// buy
			amountHeld = tradeQuantity / trade.Price
		} else if trade.Type == 2 {
			b.PnL += (amountHeld * trade.Price) - tradeQuantity
			fmt.Printf("PnL %f\n", b.PnL)
			amountHeld = 0
		}
	}
}
