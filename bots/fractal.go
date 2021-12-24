package bots

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/plot/plotter"

	"binance-bot-test/calcs"
	"binance-bot-test/display"
	candles "binance-bot-test/storage"
)

type FractalBot struct {
	PnL          float64
	Timestep     time.Duration
	RsiBuyLevel  float64
	RsiSellLevel float64

	inTrade       bool
	buySignal     bool
	entryPrice    float64
	stopLossPrice float64
	lastTradeID   int64
	hitStopLoss   bool
	allLinesValid bool

	rsiGains   []float64
	rsiLosses  []float64
	rsiAvgGain float64
	rsiAvgLoss float64
	currentRsi float64

	jawSMA5   *calcs.MovingAverage
	teethSMA8 *calcs.MovingAverage
	lipsSMA13 *calcs.MovingAverage

	fractalTopCount   int64
	runningTradeCount int64
	Trades            []Trade

	candles []*candles.Candle

	chart *display.Chart

	jawLine      *display.ChartLine
	teethLine    *display.ChartLine
	buyBars      *display.ChartLine
	sellBars     *display.ChartBar
	lipsLine     *display.ChartLine
	rsiLine      *display.ChartLine
	chartCandles *display.ChartCandles
}

func (f *FractalBot) Init(writeChart bool) {

	f.inTrade = false
	f.fractalTopCount = 0
	f.lastTradeID = -1
	f.hitStopLoss = false
	f.allLinesValid = false

	f.jawSMA5 = calcs.CreateMovingAverage(5)
	f.teethSMA8 = calcs.CreateMovingAverage(8)
	f.lipsSMA13 = calcs.CreateMovingAverage(13)

	f.rsiAvgGain = float64(math.Inf(1))
	f.rsiAvgLoss = float64(math.Inf(1))
	f.currentRsi = 100.0

	if writeChart {
		f.chart = &display.Chart{}
		f.chart.Init()

		f.jawLine = f.chart.AddLine(255, 0, 0, 1)
		f.teethLine = f.chart.AddLine(0, 255, 0, 1)
		f.lipsLine = f.chart.AddLine(0, 0, 255, 1)
		f.buyBars = f.chart.AddLine(0, 155, 155, 2)
		f.chartCandles = f.chart.AddCandles()

		f.rsiLine = f.chart.AddRsi()
	}
}

func (f *FractalBot) AddMarketTrade(trade *binance.AggTrade) {

	if trade.AggTradeID <= f.lastTradeID {
		return
	}
	price, _ := strconv.ParseFloat(trade.Price, 64)
	f.lastTradeID = trade.AggTradeID

	if f.allLinesValid {
		if f.inTrade {
			/*if price < f.stopLossPrice {
				fmt.Printf("**Hit Stop Loss: %f\n", f.stopLossPrice)
				f.inTrade = false
				f.Trades = append(f.Trades, Trade{Price: price, Timestamp: trade.Timestamp, Type: 2})
				f.buyBars.PointsChannel <- plotter.XY{X: f.candles[len(f.candles)-1].Timestamp / float64(time.Microsecond), Y: price}
			} else {
				f.stopLossPrice = math.Max(f.stopLossPrice, price-(stopLossPercent*price))
			}*/

			if f.currentRsi > f.RsiSellLevel {
				//if (((price - f.Trades[len(f.Trades)-1].Price) / f.Trades[len(f.Trades)-1].Price) > minProfitPercent) && (f.currentRsi > 70.0) {
				f.inTrade = false
				f.Trades = append(f.Trades, Trade{Price: price, Timestamp: trade.Timestamp, Type: 2})
				//fmt.Printf("Sell at %s\n", trade.Price)
			}
		}
	}

	if len(f.candles) == 0 {
		f.candles = append(f.candles, &candles.Candle{})
	}
	if ok := f.candles[len(f.candles)-1].AddTrade(trade, f.Timestep); !ok {
		// Candle has closed
		closePrice := f.candles[len(f.candles)-1].Close
		jawMA, jawMAValid := f.jawSMA5.Get(closePrice)
		if jawMAValid && (f.jawLine != nil) {
			f.jawLine.PointsChannel <- plotter.XY{X: f.candles[len(f.candles)-1].Timestamp / float64(time.Microsecond), Y: jawMA}
		}

		teethMA, teethMAValid := f.teethSMA8.Get(closePrice)
		if teethMAValid && (f.teethLine != nil) {
			f.teethLine.PointsChannel <- plotter.XY{X: f.candles[len(f.candles)-1].Timestamp / float64(time.Microsecond), Y: teethMA}
		}

		lipsMA, lipsMAValid := f.lipsSMA13.Get(closePrice)
		if lipsMAValid && (f.lipsLine != nil) {
			f.lipsLine.PointsChannel <- plotter.XY{X: f.candles[len(f.candles)-1].Timestamp / float64(time.Microsecond), Y: lipsMA}
		}
		f.allLinesValid = jawMAValid && teethMAValid && lipsMAValid
		//if (jawMA > teethMA) && (teethMA > lipsMA) {
		if true {
			if !f.inTrade && (price > teethMA) && (f.currentRsi < f.RsiBuyLevel) {
				//if !f.inTrade && (f.currentRsi < 30.0) {
				f.inTrade = true
				f.Trades = append(f.Trades, Trade{Price: price, Timestamp: trade.Timestamp, Type: 1})
				f.stopLossPrice = price - (stopLossPercent * price)
				//fmt.Printf("Buy at %s\n", trade.Price)

				if f.buyBars != nil {
					f.buyBars.PointsChannel <- plotter.XY{X: float64(trade.Timestamp) / float64(time.Microsecond), Y: price}
				}
			}
		} else {
			/*if f.inTrade {
				// Sell signal
				//if ((price - f.Trades[len(f.Trades)-1].Price) / f.Trades[len(f.Trades)-1].Price) > minProfitPercent {
				f.inTrade = false
				f.Trades = append(f.Trades, Trade{Price: price, Timestamp: trade.Timestamp, Type: 2})
				//f.sellBars.PointsChannel <- price
				f.buyBars.PointsChannel <- plotter.XY{X: float64(trade.Timestamp) / float64(time.Microsecond), Y: price}
				fmt.Printf("Sell at %s\n", trade.Price)
				//	}
			}*/
		}

		if len(f.candles) == 3 {
			// Remove the oldest candle
			if f.chartCandles != nil {
				f.chartCandles.CandlesChannel <- f.candles[0]
			}
			f.candles = append(f.candles[1:], &candles.Candle{})
		} else {
			f.candles = append(f.candles, &candles.Candle{})
			if ok := f.candles[len(f.candles)-1].AddTrade(trade, f.Timestep); !ok {
				fmt.Println("Somethings gone badly wrong")
			}
		}

		if len(f.candles) > 1 {
			difference := closePrice - f.candles[0].Close
			gain := 0.0
			loss := 0.0
			if difference > 0.0 {
				gain = difference
			} else {
				loss = math.Abs(difference)
			}

			f.rsiGains = append(f.rsiGains, gain)
			f.rsiLosses = append(f.rsiLosses, loss)

			if len(f.rsiGains) == 15 {
				// Calculate MA for gains&losses
				f.rsiAvgGain = floats.Sum(f.rsiGains) / float64(len(f.rsiGains))
				f.rsiAvgLoss = floats.Sum(f.rsiLosses) / float64(len(f.rsiLosses))

				rs := f.rsiAvgGain / f.rsiAvgLoss
				rsi := 100 - (100 / (1 + rs))

				if f.rsiLine != nil {
					f.rsiLine.PointsChannel <- plotter.XY{X: f.candles[1].Timestamp / float64(time.Microsecond), Y: rsi}
				}
				f.rsiGains = f.rsiGains[1:]
				f.rsiLosses = f.rsiLosses[1:]
				f.currentRsi = rsi
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
	if f.chart != nil {
		f.chart.Output()
	}
}

func (f *FractalBot) GetPnL(tradeQuantity float64) float64 {
	f.PnL = tradeQuantity
	amountHeld := 0.0
	for _, trade := range f.Trades {
		if trade.Type == 1 {
			// buy
			amountHeld = f.PnL / trade.Price
		} else if trade.Type == 2 {
			f.PnL = (amountHeld * trade.Price)
			amountHeld = 0
		}
	}
	return f.PnL
}
