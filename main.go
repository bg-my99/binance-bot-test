package main

import (
	"fmt"
	"time"

	"github.com/adshao/go-binance/v2"

	"binance-bot-test/config"
)

func main() {

	var cfg config.Config
	config.ReadEnv(&cfg)
	fmt.Printf("%+v", cfg)

	binance.UseTestnet = true
	wsDepthHandler := func(event *binance.WsDepthEvent) {
		fmt.Println(event)
	}
	errHandler := func(err error) {
		fmt.Println(err)
	}
	doneC, stopC, err := binance.WsDepthServe("LTCBTC", wsDepthHandler, errHandler)
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
