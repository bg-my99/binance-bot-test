package bots

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"gonum.org/v1/plot/plotter"

	"binance-bot-test/display"
	candles "binance-bot-test/storage"
)

type FractalBot struct {
	PnL           float64
	inTrade       bool
	buySignal     bool
	entryPrice    float64
	stopLossPrice float64
	lastTradeID   int64
	hitStopLoss   bool
	allLinesValid bool

	jawSMA5   []float64
	teethSMA8 []float64
	lipsSMA13 []float64

	runningJawSMA5   float64
	runningTeethSMA8 float64
	runningLipsSMA13 float64

	fractalTopCount   int64
	runningTradeCount int64
	Trades            []Trade

	candles []*candles.Candle

	chart display.Chart

	jawLine      *display.ChartLine
	teethLine    *display.ChartLine
	buyBars      *display.ChartLine
	sellBars     *display.ChartBar
	lipsLine     *display.ChartLine
	chartCandles *display.ChartCandles
}

func (f *FractalBot) Init() {

	f.inTrade = false
	f.fractalTopCount = 0
	f.lastTradeID = -1
	f.hitStopLoss = false
	f.allLinesValid = false

	f.chart = display.Chart{}
	f.chart.Init()

	f.jawLine = f.chart.AddLine(255, 0, 0, 1)
	f.teethLine = f.chart.AddLine(0, 255, 0, 1)
	f.lipsLine = f.chart.AddLine(0, 0, 255, 1)
	f.buyBars = f.chart.AddLine(0, 155, 155, 2)
	f.chartCandles = f.chart.AddCandles()
}

func (f *FractalBot) AddMarketTrade(trade *binance.AggTrade) {

	if trade.AggTradeID <= f.lastTradeID {
		return
	}
	price, _ := strconv.ParseFloat(trade.Price, 64)
	f.lastTradeID = trade.AggTradeID

	if f.allLinesValid {
		if f.inTrade {
			if price < f.stopLossPrice {
				fmt.Printf("**Hit Stop Loss: %f\n", f.stopLossPrice)
				f.inTrade = false
				f.Trades = append(f.Trades, Trade{Price: price, Timestamp: trade.Timestamp, Type: 2})
				f.buyBars.PointsChannel <- plotter.XY{X: f.candles[len(f.candles)-1].Timestamp / float64(time.Microsecond), Y: price}
			} else {
				f.stopLossPrice = math.Max(f.stopLossPrice, price-(stopLossPercent*price))
			}

			if ((price - f.Trades[len(f.Trades)-1].Price) / f.Trades[len(f.Trades)-1].Price) > minProfitPercent {
				f.inTrade = false
				f.Trades = append(f.Trades, Trade{Price: price, Timestamp: trade.Timestamp, Type: 2})
				fmt.Printf("Sell at %s\n", trade.Price)
			}
		}

		if (f.runningJawSMA5 > f.runningTeethSMA8) && (f.runningTeethSMA8 > f.runningLipsSMA13) {
			if !f.inTrade && (price > f.runningJawSMA5) {
				f.inTrade = true
				f.Trades = append(f.Trades, Trade{Price: price, Timestamp: trade.Timestamp, Type: 1})
				f.stopLossPrice = price - (stopLossPercent * price)
				//f.buyBars.PointsChannel <- price
				f.buyBars.PointsChannel <- plotter.XY{X: float64(trade.Timestamp) / float64(time.Microsecond), Y: price}
				fmt.Printf("Buy at %s\n", trade.Price)
			}
		} else {
			if f.inTrade {
				// Sell signal
				//if ((price - f.Trades[len(f.Trades)-1].Price) / f.Trades[len(f.Trades)-1].Price) > minProfitPercent {
				f.inTrade = false
				f.Trades = append(f.Trades, Trade{Price: price, Timestamp: trade.Timestamp, Type: 2})
				//f.sellBars.PointsChannel <- price
				f.buyBars.PointsChannel <- plotter.XY{X: float64(trade.Timestamp) / float64(time.Microsecond), Y: price}
				fmt.Printf("Sell at %s\n", trade.Price)
				//	}
			}
		}
	}

	if len(f.candles) == 0 {
		f.candles = append(f.candles, &candles.Candle{})
	}
	if ok := f.candles[len(f.candles)-1].AddTrade(trade); !ok {
		// Candle has closed
		closePrice := f.candles[len(f.candles)-1].Close
		if len(f.jawSMA5) < 5 {
			f.jawSMA5 = append(f.jawSMA5, closePrice)
		} else {
			f.runningJawSMA5 = 0.0
			for _, price := range f.jawSMA5 {
				f.runningJawSMA5 += price
			}
			f.runningJawSMA5 /= 5.0
			f.jawSMA5 = append(f.jawSMA5[1:], closePrice)
			f.jawLine.PointsChannel <- plotter.XY{X: f.candles[len(f.candles)-1].Timestamp / float64(time.Microsecond), Y: f.runningJawSMA5}
		}

		if len(f.teethSMA8) < 8 {
			f.teethSMA8 = append(f.teethSMA8, closePrice)
		} else {
			f.runningTeethSMA8 = 0.0
			for _, price := range f.teethSMA8 {
				f.runningTeethSMA8 += price
			}
			f.runningTeethSMA8 /= 8.0
			f.teethSMA8 = append(f.teethSMA8[1:], closePrice)
			f.teethLine.PointsChannel <- plotter.XY{X: f.candles[len(f.candles)-1].Timestamp / float64(time.Microsecond), Y: f.runningTeethSMA8}
		}

		if len(f.lipsSMA13) < 13 {
			f.lipsSMA13 = append(f.lipsSMA13, closePrice)
		} else {
			f.allLinesValid = true
			f.runningLipsSMA13 = 0.0
			for _, price := range f.lipsSMA13 {
				f.runningLipsSMA13 += price
			}
			f.runningLipsSMA13 /= 13.0
			f.lipsSMA13 = append(f.lipsSMA13[1:], closePrice)
			f.lipsLine.PointsChannel <- plotter.XY{X: f.candles[len(f.candles)-1].Timestamp / float64(time.Microsecond), Y: f.runningLipsSMA13}
		}

		if len(f.candles) == 3 {
			/*if f.buySignal {
				// See if we should cancel the signal
				if f.candles[2].Low < f.runningTeethSMA8 {
					f.buySignal = false
					f.fractalTopCount = 0
				} else {
					// Or should we buy?
					if f.candles[2].High > f.entryPrice {
						f.inTrade = true
						f.Trades = append(f.Trades, Trade{Price: price, Timestamp: trade.Timestamp, Type: 1})
						f.stopLossPrice = price - (stopLossPercent * price)
						fmt.Printf("Buy at %s\n", trade.Price)
						f.buySignal = false
					}
				}
			}
			if f.fractalTopCount > 0 {
				if f.candles[2].Low > f.runningTeethSMA8 {
					f.fractalTopCount++
					if f.fractalTopCount > 5 && !f.inTrade {
						f.buySignal = true
					}
				}
			}

			// Check for fractals
			if (f.candles[1].High > f.candles[0].High) && (f.candles[1].High > f.candles[2].High) && (f.fractalTopCount == 0) {
				f.entryPrice = f.candles[1].High
				f.fractalTopCount = 1
			} else if (f.candles[1].Low < f.candles[0].Low) && (f.candles[1].Low > f.candles[2].Low) {

			}
			*/
			// Remove the oldest candle
			f.chartCandles.CandlesChannel <- f.candles[0]
			f.candles = append(f.candles[1:], &candles.Candle{})
		} else {
			f.candles = append(f.candles, &candles.Candle{})
			if ok := f.candles[len(f.candles)-1].AddTrade(trade); !ok {
				fmt.Println("Somethings gone badly wrong")
			}
		}
	}

}

func (f *FractalBot) DisplayPnL(tradeQuantity float64) {
	f.PnL = tradeQuantity
	amountHeld := 0.0
	for _, trade := range f.Trades {
		if trade.Type == 1 {
			// buy
			amountHeld = f.PnL / trade.Price
		} else if trade.Type == 2 {
			f.PnL = (amountHeld * trade.Price)
			fmt.Printf("PnL %f\n", f.PnL)
			amountHeld = 0
		}
	}
	f.chart.Output()
}
