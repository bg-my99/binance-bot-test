package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
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

	req, err := http.NewRequest("GET", "https://api.binance.com/api/v3/aggTrades", nil)
	if err != nil {
		fmt.Print(err)
		return
	}

	yesterday := time.Now().AddDate(0, 0, -1)
	yesterday, _ = time.Parse("2006/01/02", yesterday.Format("2006/01/02"))

	q := req.URL.Query()
	q.Add("symbol", "GALABUSD")
	q.Add("startTime", strconv.FormatInt(yesterday.Add(time.Duration(-time.Minute)).UnixNano()/int64(time.Millisecond), 10))
	q.Add("endTime", strconv.FormatInt(yesterday.UnixNano()/int64(time.Millisecond), 10))
	req.URL.RawQuery = q.Encode()

	resp, err := http.Get(req.URL.String())
	if err != nil {
		fmt.Println(err)
		return
	}
	if resp.StatusCode != 200 {
		fmt.Printf("Received http %d code\n", resp.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	//Convert the body to type string
	res := make([]*binance.AggTrade, 0)
	err = json.Unmarshal(body, &res)

	if err != nil {
		fmt.Println("Couldnt parse json from initial trades")
	}
	fromID := res[0].AggTradeID

	current_time := yesterday.UnixNano() / int64(time.Millisecond)
	end_time := yesterday.AddDate(0, 0, 1).UnixNano() / int64(time.Millisecond)

	for current_time < end_time {
		req, err := http.NewRequest("GET", "https://api.binance.com/api/v3/aggTrades", nil)
		if err != nil {
			fmt.Print(err)
			return
		}

		q := req.URL.Query()
		q.Add("symbol", "GALABUSD")
		q.Add("limit", "1000")
		q.Add("fromId", strconv.FormatInt(fromID, 10))
		req.URL.RawQuery = q.Encode()

		resp, err := http.Get(req.URL.String())
		if err != nil {
			fmt.Println(err)
			return
		}
		if resp.StatusCode != 200 {
			fmt.Printf("Received http %d code\n", resp.StatusCode)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		res := make([]*binance.AggTrade, 0)
		err = json.Unmarshal(body, &res)

		if err != nil {
			fmt.Println("Couldnt parse json from initial trades")
		}

		for _, trade := range res {
			fmt.Printf("%v,%v\n", trade.Price, trade.Quantity)
		}
		time.Sleep(500 * time.Millisecond)

		fromID = res[len(res)-1].AggTradeID
		current_time = res[len(res)-1].Timestamp
	}
}
