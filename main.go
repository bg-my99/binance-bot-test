package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"

	"binance-bot-test/config"
)

func getTrades() []binance.AggTrade {
	req, err := http.NewRequest("GET", "https://api.binance.com/api/v3/aggTrades", nil)
	if err != nil {
		fmt.Print(err)
		return nil
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

	current_time := yesterday.UnixNano() / int64(time.Millisecond)
	end_time := yesterday.AddDate(0, 0, 1).UnixNano() / int64(time.Millisecond)

	trades := []binance.AggTrade{}

	for current_time < end_time {
		req, err := http.NewRequest("GET", "https://api.binance.com/api/v3/aggTrades", nil)
		if err != nil {
			fmt.Print(err)
			return nil
		}

		q := req.URL.Query()
		q.Add("symbol", "GALABUSD")
		q.Add("limit", "1000")
		q.Add("fromId", strconv.FormatInt(fromID, 10))
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

func getMovingAverage(pts plotter.XYs) plotter.XYs {
	movingAverage := plotter.XYs{}
	runningCount := 0.0
	for i, trade := range pts {
		if i >= 20 && i < (len(pts)-20) {
			movingAverage = append(movingAverage, plotter.XY{X: trade.X, Y: runningCount / 20.0})
			runningCount -= pts[i-20].Y
		}
		runningCount += trade.Y
	}
	return movingAverage
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
		trades = getTrades()
		if cfg.WriteTrades {
			file, _ := json.MarshalIndent(trades, "", " ")
			_ = ioutil.WriteFile("trades.json", file, 0644)
		}
	} else if cfg.TradesSource == "file" {
		file, _ := ioutil.ReadFile("trades.json")
		_ = json.Unmarshal([]byte(file), &trades)
	}

	p := plot.New()
	p.Title.Text = "Plotutil example"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"

	pts := plotter.XYs{}

	for _, trade := range trades {
		y, _ := strconv.ParseFloat(trade.Price, 64)
		pts = append(pts, plotter.XY{X: float64(trade.Timestamp), Y: y})
	}
	l, err := plotter.NewLine(pts)
	if err != nil {
		panic(err)
	}
	l.LineStyle.Width = vg.Points(1)
	l.LineStyle.Color = color.RGBA{B: 255, A: 255}
	p.Add(l)

	movingAverage := getMovingAverage(pts)
	m, err := plotter.NewLine(movingAverage)
	if err != nil {
		panic(err)
	}
	m.LineStyle.Width = vg.Points(1)
	m.LineStyle.Color = color.RGBA{R: 200, A: 128}
	p.Add(m)

	// Save the plot to a PNG file.
	if err := p.Save(32*vg.Inch, 16*vg.Inch, "points.png"); err != nil {
		panic(err)
	}
}
