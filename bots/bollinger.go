package bots

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"

	"binance-bot-test/calcs"
	candles "binance-bot-test/storage"
)

const numPoints = 20
const stopLossPercent = 0.01
const minProfitPercent = 0.03

type BollingerBot struct {
	PnL           float64
	inTrade       bool
	stopLossPrice float64
	lastTradeID   int64
	hitStopLoss   bool

	runningEMA20  float64
	runningEMA50  float64
	runningEMA100 float64

	runningTradeCount int64
	Trades            []Trade
	candles           candles.Candles

	candle candles.Candle
}

func (b *BollingerBot) Init() {

	b.candles = candles.Candles{}
	b.candles.Init(0, int64((time.Second*60)/time.Millisecond))
	b.inTrade = false
	b.lastTradeID = -1
	b.hitStopLoss = false
}

func (b *BollingerBot) AddMarketTrade(trade *binance.AggTrade) {

	if trade.AggTradeID <= b.lastTradeID {
		return
	}

	b.candles.AddTrade(trade)

	pts, _ := b.candles.GetSortedCandles()
	movingAverage := calcs.GetMovingAverage(pts, numPoints)
	standardDeviation := calcs.GetStandardDeviation(pts, movingAverage, numPoints)

	if len(movingAverage) > 0 {
		price, _ := strconv.ParseFloat(trade.Price, 64)
		//fmt.Printf("p:%f\n", price)

		if b.inTrade {
			// Check stop loss
			if price < b.stopLossPrice {
				fmt.Printf("**Hit Stop Loss: %f\n", b.stopLossPrice)
				b.inTrade = false
				b.hitStopLoss = true
				b.Trades = append(b.Trades, Trade{Price: price, Timestamp: trade.Timestamp, Type: 2})
			} else {
				b.stopLossPrice = math.Max(b.stopLossPrice, price-(stopLossPercent*price))
				//fmt.Printf("slp: %f\n", b.stopLossPrice)
			}
		}
		if price > (movingAverage[movingAverage.Len()-1].Y + (2.2 * standardDeviation[standardDeviation.Len()-1].Y)) {
			if b.inTrade && (len(b.Trades) > 0) {
				if ((price - b.Trades[len(b.Trades)-1].Price) / b.Trades[len(b.Trades)-1].Price) > minProfitPercent {
					b.inTrade = false
					b.Trades = append(b.Trades, Trade{Price: price, Timestamp: trade.Timestamp, Type: 2})
					fmt.Printf("Sell at %s\n", trade.Price)
				}
			}
		} else if price < (movingAverage[movingAverage.Len()-1].Y - (2.2 * standardDeviation[standardDeviation.Len()-1].Y)) {
			if !b.inTrade {
				ema20, ema50, _ := calcs.GetExpMovingAverages(movingAverage, numPoints)
				if ema20[ema20.Len()-1].Y > ema50[ema50.Len()-1].Y {
					b.inTrade = true
					b.Trades = append(b.Trades, Trade{Price: price, Timestamp: trade.Timestamp, Type: 1})
					b.stopLossPrice = price - (stopLossPercent * price)
					fmt.Printf("Buy at %s\n", trade.Price)
				}
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
			amountHeld = b.PnL / trade.Price
		} else if trade.Type == 2 {
			b.PnL = (amountHeld * trade.Price)
			fmt.Printf("PnL %f\n", b.PnL)
			amountHeld = 0
		}
	}
}
