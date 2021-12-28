package bots

import (
	"sync"
	"time"

	"github.com/adshao/go-binance/v2"
)

type BotRun struct {
	rsiBuyPrice  float64
	rsiSellPrice float64

	movingAveragePeriod float64
	timeStep            float64
}

type BotRunner struct {
	RsiBuyPriceMin  float64
	RsiBuyPriceMax  float64
	RsiBuyPriceStep float64

	RsiSellPriceMin  float64
	RsiSellPriceMax  float64
	RsiSellPriceStep float64

	MovingAveragePeriodMin  float64
	MovingAveragePeriodMax  float64
	MovingAveragePeriodStep float64

	TimeSteps []time.Duration
}

type BotRunResult struct {
	Timestep time.Duration
	RsiBuy   float64
	RsiSell  float64
	PnL      float64
	Trades   *[]Trade
}

func (br *BotRunner) Run(trades []binance.AggTrade, chResults chan BotRunResult) {
	var wg sync.WaitGroup

	for _, timestep := range br.TimeSteps {
		for rsiSellCurrent := br.RsiSellPriceMin; rsiSellCurrent < br.RsiSellPriceMax; rsiSellCurrent += br.RsiSellPriceStep {
			for rsiBuyCurrent := br.RsiBuyPriceMin; rsiBuyCurrent < br.RsiBuyPriceMax; rsiBuyCurrent += br.RsiBuyPriceStep {
				wg.Add(1)
				go func(ts time.Duration, rsiBuy float64, rsiSell float64, wg *sync.WaitGroup) {
					bot := FractalBot{Timestep: ts, RsiBuyLevel: rsiBuy, RsiSellLevel: rsiSell}
					bot.Init(false)

					for _, trade := range trades {
						bot.AddMarketTrade(&trade)
					}
					//fmt.Printf("PnL for timestep %v, rsiB:%f, rsiS:%f:%f\n", ts, rsiBuy, rsiSell, bot.GetPnL(100))
					chResults <- BotRunResult{Timestep: ts, RsiBuy: rsiBuy, RsiSell: rsiSell, PnL: bot.GetPnL(100), Trades: &bot.Trades}
					defer wg.Done()

				}(timestep, rsiBuyCurrent, rsiSellCurrent, &wg)
			}
		}
	}
	wg.Wait()
}
