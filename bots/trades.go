package bots

import "github.com/adshao/go-binance/v2"

type Trade struct {
	Price     float64
	Quantity  float64
	Timestamp int64
	Type      uint32

	trades []binance.AggTrade
}
