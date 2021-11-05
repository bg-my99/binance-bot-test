package candles

import (
	"strconv"

	"github.com/adshao/go-binance/v2"
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
	index := trade.Timestamp % c.timeStep
	if _, ok := c.candles[index]; !ok {
		c.candles[index] = Candle{trades: make([]binance.AggTrade, 0)}
	}
	var candle = c.candles[index]
	candle.trades = append(c.candles[index].trades, *trade)

	// Recalc values for candle
	totalQuantity := 0.0
	totalPrice := 0.0
	for _, trade := range candle.trades {
		price, _ := strconv.ParseFloat(trade.Price, 64)
		quantity, _ := strconv.ParseInt(trade.Price, 10, 64)
		totalPrice += (price * float64(quantity))
		totalQuantity += float64(quantity)
	}
	candle.WeightedAverage = totalPrice / totalQuantity
}
