package candles

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/pplcc/plotext/custplotter"
	"gonum.org/v1/plot/plotter"
)

type Candle struct {
	Open            float64
	Close           float64
	High            float64
	Low             float64
	Volume          float64
	Timestamp       float64
	WeightedAverage float64

	trades []binance.AggTrade
}

func (c *Candle) WouldCloseWithTrade(trade *binance.AggTrade, timestep time.Duration) bool {
	//return (trade.Timestamp-int64(c.Timestamp) > int64((time.Second*600)/time.Millisecond))
	return (trade.Timestamp-int64(c.Timestamp) > int64(timestep/time.Millisecond))
}

func (c *Candle) AddTrade(trade *binance.AggTrade, timestep time.Duration) bool {

	price, _ := strconv.ParseFloat(trade.Price, 64)
	quantity, _ := strconv.ParseFloat(trade.Quantity, 64)

	if len(c.trades) == 0 {
		c.Open = price
		c.High = price
		c.Low = price
		c.Timestamp = float64(trade.Timestamp)
	} else {
		if trade.Timestamp < c.trades[len(c.trades)-1].Timestamp {
			fmt.Println("Trade with dodgy timestamp")
			return false
		}
		if c.WouldCloseWithTrade(trade, timestep) {
			closePrice, _ := strconv.ParseFloat(trade.Price, 64)
			c.Close = closePrice
			return false
		}
		c.Low = math.Min(price, c.Low)
		c.High = math.Max(price, c.High)
	}
	c.Volume += quantity
	c.trades = append(c.trades, *trade)
	return true
}

type Candles struct {
	timeStep   int64
	maxCandles int64
	baseIndex  int64
	candles    []*Candle
}

func (c *Candles) Init(maxCandles int64, timeStep int64) {
	c.timeStep = timeStep
	c.maxCandles = maxCandles
}

/*func (c *Candles) AddTrade(trade *binance.AggTrade) {
	index := trade.Timestamp - (trade.Timestamp % c.timeStep)
	if len(c.candles) == 0 {
		c.baseIndex = index
	}
	candleIndex := (index - c.baseIndex) / c.timeStep
	if candleIndex > int64(len(c.candles)-1) {
		c.candles = append(c.candles, &Candle{Timestamp: float64(trade.Timestamp), trades: make([]binance.AggTrade, 0)})
	}
	c.candles[candleIndex].AddTrade(trade, timestep)
}*/

func (c *Candles) GetSortedCandles() (plotter.XYs, custplotter.TOHLCVs) {

	candlesToReturn := plotter.XYs{}
	candlesToReturn2 := make(custplotter.TOHLCVs, len(c.candles))
	for i, candle := range c.candles {
		t := float64(candle.Timestamp) / float64(time.Microsecond)
		candlesToReturn = append(candlesToReturn, plotter.XY{X: t, Y: candle.WeightedAverage})

		candlesToReturn2[i].T = t
		candlesToReturn2[i].O = candle.Open
		candlesToReturn2[i].H = candle.High
		candlesToReturn2[i].L = candle.Low
		candlesToReturn2[i].C = candle.Close
		candlesToReturn2[i].V = candle.Volume
	}
	return candlesToReturn, candlesToReturn2
}

func GetHLCVs(candles []*Candle) custplotter.TOHLCVs {

	hlcvs := make(custplotter.TOHLCVs, len(candles))
	for i, candle := range candles {
		t := float64(candle.Timestamp) / float64(time.Microsecond)

		hlcvs[i].T = t
		hlcvs[i].O = candle.Open
		hlcvs[i].H = candle.High
		hlcvs[i].L = candle.Low
		hlcvs[i].C = candle.Close
		hlcvs[i].V = candle.Volume
	}
	return hlcvs
}
