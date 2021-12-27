package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"

	"binance-bot-test/bots"
	"binance-bot-test/config"
)

const numPoints = 20

func getTrades(date time.Time) []binance.AggTrade {
	req, err := http.NewRequest("GET", "https://api.binance.com/api/v3/aggTrades", nil)
	if err != nil {
		fmt.Print(err)
		return nil
	}

	// Strip off the time
	day, _ := time.Parse("2006/01/02", date.Format("2006/01/02"))

	symbol := "FTMBUSD"
	q := req.URL.Query()
	q.Add("symbol", symbol)
	q.Add("startTime", strconv.FormatInt(day.Add(time.Duration(-time.Minute)).UnixNano()/int64(time.Millisecond), 10))
	q.Add("endTime", strconv.FormatInt(day.UnixNano()/int64(time.Millisecond), 10))
	req.URL.RawQuery = q.Encode()

	resp, err := http.Get(req.URL.String())
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if resp.StatusCode != 200 {
		fmt.Printf("Received http %d code\n", resp.StatusCode)
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	//Convert the body to type string
	res := make([]*binance.AggTrade, 0)
	err = json.Unmarshal(body, &res)

	if err != nil {
		fmt.Println("Couldnt parse json from initial trades")
	}
	fromID := res[0].AggTradeID

	fmt.Println("Fetching for " + day.Format("2006-01-02"))

	current_time := day.UnixNano() / int64(time.Millisecond)
	end_time := day.AddDate(0, 0, 1).UnixNano() / int64(time.Millisecond)

	trades := []binance.AggTrade{}

	for current_time < end_time {
		req, err := http.NewRequest("GET", "https://api.binance.com/api/v3/aggTrades", nil)
		if err != nil {
			fmt.Print("NewRequest returned:", err)
			return nil
		}

		q := req.URL.Query()
		q.Add("symbol", symbol)
		q.Add("limit", "1000")
		q.Add("fromId", strconv.FormatInt(fromID, 10))
		req.URL.RawQuery = q.Encode()

		resp, err := http.Get(req.URL.String())
		if err != nil {
			fmt.Println("Get returned:", err)
			return nil
		}
		if resp.StatusCode != 200 {
			fmt.Printf("Received http %d code\n", resp.StatusCode)
			return nil
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("ReadAll returned:", err)
			return nil
		}
		res := make([]*binance.AggTrade, 0)
		err = json.Unmarshal(body, &res)

		if err != nil {
			fmt.Println("Couldnt parse json from initial trades")
		}

		for _, trade := range res {
			trades = append(trades, *trade)
		}
		time.Sleep(500 * time.Millisecond)

		fromID = res[len(res)-1].AggTradeID
		current_time = res[len(res)-1].Timestamp
	}
	return trades
}

type ValuePoint struct {
	RsiSell float64 `json:"rsiSell"`
	PnL     float64 `json:"pnl"`
}

type ValueLine struct {
	Points []ValuePoint `json:"points"`
}

type ValueLines struct {
	Lines map[string]ValueLine `json:"lines"`
}

type RunResults struct {
	LinesByTimestamp map[time.Duration]ValueLines `json:"linesByTimestamp"`
}

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

	var trades []binance.AggTrade
	if cfg.TradesSource == "binance" {
		date := time.Now()
		if cfg.FetchForDate != "" {
			layout := "2006-01-02"
			t, err := time.Parse(layout, cfg.FetchForDate)
			if err != nil {
				fmt.Println("Could not parse date from:" + cfg.FetchForDate)
			} else {
				date = t
				currentYear, currentMonth, _ := date.Date()
				currentLocation := date.Location()

				firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
				currentDate := firstOfMonth
				lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

				for ; !currentDate.After(lastOfMonth) && currentDate.Before(time.Now()); currentDate = currentDate.AddDate(0, 0, 1) {
					trades = getTrades(currentDate)
					filename := fmt.Sprintf("trades-%s.json", currentDate.Format("2006-01-02"))
					file, _ := json.MarshalIndent(trades, "", " ")
					_ = ioutil.WriteFile(filename, file, 0644)
					fmt.Println("Wrote:", filename)
				}
			}
		} else {
			trades = getTrades(date)
			if cfg.WriteTrades {
				file, _ := json.MarshalIndent(trades, "", " ")
				_ = ioutil.WriteFile("trades.json", file, 0644)
			}
		}
		return
	} else if cfg.TradesSource == "file" {
		filenames := []string{
			/*"data/trades-2021-11-01.json", "data/trades-2021-11-02.json",
			"data/trades-2021-11-03.json", "data/trades-2021-11-04.json",
			"data/trades-2021-11-05.json", "data/trades-2021-11-06.json",
			"data/trades-2021-11-07.json", "data/trades-2021-11-08.json",
			"data/trades-2021-11-09.json", "data/trades-2021-11-10.json",
			"data/trades-2021-11-11.json", "data/trades-2021-11-12.json",
			"data/trades-2021-11-13.json", "data/trades-2021-11-14.json",
			"data/trades-2021-11-15.json", "data/trades-2021-11-16.json",*/
			"data/trades-2021-11-18.json", "data/trades-2021-11-19.json",
			//"data/trades-2021-11-20.json", "data/trades-2021-11-21.json",
			//"data/trades-2021-11-22.json", "data/trades-2021-11-23.json",
		}
		var tradesFromFile []binance.AggTrade
		for _, filename := range filenames {
			file, _ := ioutil.ReadFile(filename)
			_ = json.Unmarshal([]byte(file), &tradesFromFile)
			trades = append(trades, tradesFromFile...)
		}
	}

	botrunner := bots.BotRunner{
		TimeSteps:      []time.Duration{time.Second * 60, time.Second * 120, time.Second * 300, time.Second * 600, time.Second * 900},
		RsiBuyPriceMin: 10.0, RsiBuyPriceMax: 40.0, RsiBuyPriceStep: 4.0,
		RsiSellPriceMin: 60.0, RsiSellPriceMax: 90.0, RsiSellPriceStep: 5.0,
	}
	resultsChannel := make(chan bots.BotRunResult)
	results := RunResults{LinesByTimestamp: make(map[time.Duration]ValueLines)}

	go func(chResults chan bots.BotRunResult) {
		for result := range chResults {
			//fmt.Printf("PnL for timestep %v, rsiB:%f, rsiS:%f:%f\n", result.Timestep, result.RsiBuy, result.RsiSell, result.PnL)

			if _, ok := results.LinesByTimestamp[result.Timestep]; !ok {
				results.LinesByTimestamp[result.Timestep] = ValueLines{Lines: make(map[string]ValueLine)}
			}
			rsiAsStr := fmt.Sprintf("%f", result.RsiBuy)
			if _, ok := results.LinesByTimestamp[result.Timestep].Lines[rsiAsStr]; !ok {
				results.LinesByTimestamp[result.Timestep].Lines[rsiAsStr] = ValueLine{Points: make([]ValuePoint, 0)}
			}
			newPoints := append(results.LinesByTimestamp[result.Timestep].Lines[rsiAsStr].Points, ValuePoint{RsiSell: result.RsiSell, PnL: result.PnL})
			results.LinesByTimestamp[result.Timestep].Lines[rsiAsStr] = ValueLine{Points: newPoints}
		}
	}(resultsChannel)

	botrunner.Run(trades, resultsChannel)

	prettyJSON, error := json.MarshalIndent(results, "", "\t")
	if error != nil {
		fmt.Println("JSON parse error: ", error)
		return
	}
	fmt.Println(string(prettyJSON))
}
