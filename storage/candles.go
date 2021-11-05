package candles

import (
	"math"
	"sort"
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
	WeightedAverage float64

	trades []binance.AggTrade
}

type Candles struct {
	timeStep int64
	candles  map[int64]Candle
}

func (c *Candles) Init(timeStep int64) {
	c.timeStep = timeStep
	c.candles = make(map[int64]Candle)
}

func (c *Candles) AddTrade(trade *binance.AggTrade) {
	index := trade.Timestamp - (trade.Timestamp % c.timeStep)
	if _, ok := c.candles[index]; !ok {
		c.candles[index] = Candle{trades: make([]binance.AggTrade, 0)}
	}
	var candle = c.candles[index]
	candle.trades = append(c.candles[index].trades, *trade)

	// Recalc values for candle
	totalQuantity := 0.0
	minPrice := math.Inf(1)
	maxPrice := math.Inf(-1)
	openTime := index + c.timeStep
	openPrice := 0.0
	closeTime := index - c.timeStep
	closePrice := 0.0
	totalPrice := 0.0

	for _, t := range candle.trades {
		price, _ := strconv.ParseFloat(t.Price, 64)
		quantity, _ := strconv.ParseFloat(t.Quantity, 64)

		if t.Timestamp > closeTime {
			// new close price
			closePrice = price
			closeTime = t.Timestamp
		}
		if t.Timestamp < openTime {
			// new open price
			openPrice = price
			openTime = t.Timestamp
		}
		minPrice = math.Min(price, minPrice)
		maxPrice = math.Max(price, maxPrice)

		totalPrice += (price * float64(quantity))
		totalQuantity += float64(quantity)
	}
	if totalQuantity != 0.0 {
		candle.Open = openPrice
		candle.Close = closePrice
		candle.High = maxPrice
		candle.Low = minPrice
		candle.Volume = totalQuantity
		candle.WeightedAverage = totalPrice / totalQuantity
		c.candles[index] = candle
	}
}

func (c *Candles) GetSortedCandles() (plotter.XYs, custplotter.TOHLCVs) {

	keys := make([]int64, 0, len(c.candles))
	for k := range c.candles {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] <= keys[j] })

	candlesToReturn := plotter.XYs{}
	candlesToReturn2 := make(custplotter.TOHLCVs, len(keys))
	for i, k := range keys {
		candle := c.candles[k]
		t := float64(k) / float64(time.Microsecond)
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
