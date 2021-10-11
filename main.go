package main

import (
	"context"
	"fmt"
	"time"

	"github.com/adshao/go-binance/v2"

	"binance-bot-test/config"
)

func main() {

	var cfg config.Config
	config.ReadEnv(&cfg)
	fmt.Printf("%+v\n", cfg)

	whichNet := "main network"
	if cfg.UseTestnet {
		whichNet = "testnet"
	}
	fmt.Printf("Connecting to %s\n", whichNet)
	binance.UseTestnet = cfg.UseTestnet

	client := binance.NewClient(cfg.AccessKeys.ApiKey, cfg.AccessKeys.SecretKey)
	symbol := "LTCBTC"
	depthSnapshot, err := client.NewDepthService().Symbol(symbol).Limit(1000).Do(context.Background())
	fmt.Printf("%v\n", depthSnapshot)

	wsDepthHandler := func(event *binance.WsDepthEvent) {

		fmt.Printf("%s:\n", time.Unix(event.Time/1e3, (event.Time%1e3)*1e6).Format(time.RFC3339))

		for _, bid := range event.Bids {
			fmt.Printf("\t%v\n", bid)
		}

		for _, ask := range event.Asks {
			fmt.Printf("\t%v\n", ask)
		}
	}
	errHandler := func(err error) {
		fmt.Println(err)
	}
	doneC, stopC, err := binance.WsDepthServe(symbol, wsDepthHandler, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	// use stopC to exit
	go func() {
		time.Sleep(5 * time.Second)
		stopC <- struct{}{}
	}()
	// remove this if you do not want to be blocked here
	<-doneC
}
