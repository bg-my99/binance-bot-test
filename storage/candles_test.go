package candles

import (
	"testing"
	"time"

	"github.com/adshao/go-binance/v2"
)

func TestCandlesTimePeriod(t *testing.T) {
	var candles = Candles{}
	candles.Init(int64(time.Second * 300))

	tm := time.Now().UnixNano() / int64(time.Millisecond)
	tm_10minsago := time.Now().Add(time.Duration(-10)*time.Minute).UnixNano() / int64(time.Millisecond)

	candles.AddTrade(&binance.AggTrade{Price: "20", Quantity: "10", Timestamp: tm})
	candles.AddTrade(&binance.AggTrade{Price: "40", Quantity: "20", Timestamp: tm_10minsago})

	if len(candles.candles) != 2 {
		t.Errorf("Expected 2 items in candles.candles, got %d\n", len(candles.candles))
	}
}

func TestCandlesAverage(t *testing.T) {
	var candles = Candles{}
	candles.Init(int64(time.Second * 300))

	tm := time.Now().UnixNano() / int64(time.Millisecond)

	candles.AddTrade(&binance.AggTrade{Price: "20", Quantity: "15", Timestamp: tm})
	candles.AddTrade(&binance.AggTrade{Price: "40", Quantity: "25", Timestamp: tm})

	if len(candles.candles) != 1 {
		t.Errorf("Expected single candle in candles.candles, got %d\n", len(candles.candles))
	}
	for _, candle := range candles.candles {
		if candle.WeightedAverage != 32.5 {
			t.Errorf("Weighted Average should be 32.5, got %f\n", candle.WeightedAverage)
		}
	}
}
