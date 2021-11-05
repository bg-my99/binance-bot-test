package candles

import (
	"sort"
	"strconv"

	"github.com/adshao/go-binance/v2"
	"gonum.org/v1/plot/plotter"
)

type Candle struct {
	Open            float64
	Close           float64
	High            float64
	Low             float64
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
	totalPrice := 0.0

	for _, t := range candle.trades {
		price, _ := strconv.ParseFloat(t.Price, 64)
		quantity, _ := strconv.ParseFloat(t.Quantity, 64)

		totalPrice += (price * float64(quantity))
		totalQuantity += float64(quantity)
	}
	if totalQuantity != 0.0 {
		candle.WeightedAverage = totalPrice / totalQuantity
		c.candles[index] = candle
	}
}

func (c *Candles) GetSortedCandles() plotter.XYs {

	keys := make([]int64, 0, len(c.candles))
	for k := range c.candles {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] <= keys[j] })

	candlesToReturn := plotter.XYs{}
	for _, k := range keys {
		candle := c.candles[k]
		candlesToReturn = append(candlesToReturn, plotter.XY{X: float64(k), Y: candle.WeightedAverage})
	}
	return candlesToReturn
}
