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
	"github.com/pplcc/plotext/custplotter"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"

	"binance-bot-test/calcs"
	"binance-bot-test/config"
	candles "binance-bot-test/storage"
)

const numPoints = 20

func getTrades(date time.Time) []binance.AggTrade {
	req, err := http.NewRequest("GET", "https://api.binance.com/api/v3/aggTrades", nil)
	if err != nil {
		fmt.Print(err)
		return nil
	}

	previousDay := date.AddDate(0, 0, -1)
	previousDay, _ = time.Parse("2006/01/02", previousDay.Format("2006/01/02"))

	q := req.URL.Query()
	q.Add("symbol", "GALABUSD")
	q.Add("startTime", strconv.FormatInt(previousDay.Add(time.Duration(-time.Minute)).UnixNano()/int64(time.Millisecond), 10))
	q.Add("endTime", strconv.FormatInt(previousDay.UnixNano()/int64(time.Millisecond), 10))
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

	fmt.Println("Fetching for " + previousDay.Format("2006-01-02"))

	current_time := previousDay.UnixNano() / int64(time.Millisecond)
	end_time := previousDay.AddDate(0, 0, 1).UnixNano() / int64(time.Millisecond)

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
			}
		}
		trades = getTrades(date)
		if cfg.WriteTrades {
			file, _ := json.MarshalIndent(trades, "", " ")
			_ = ioutil.WriteFile("trades.json", file, 0644)
		}
	} else if cfg.TradesSource == "file" {
		file, _ := ioutil.ReadFile("trades.json")
		_ = json.Unmarshal([]byte(file), &trades)
	}

	candles := candles.Candles{}
	candles.Init(int64((time.Second * 300) / time.Millisecond))

	for _, trade := range trades {
		candles.AddTrade(&trade)
	}
	p := plot.New()
	p.Title.Text = "Plotutil example"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"

	xticks := plot.TimeTicks{Format: "2006-01-02\n15:04"}
	p.X.Tick.Marker = xticks

	pts, hlcvs := candles.GetSortedCandles()
	l, err := plotter.NewLine(pts)
	if err != nil {
		panic(err)
	}
	l.LineStyle.Width = vg.Points(1)
	l.LineStyle.Color = color.RGBA{R: 255, A: 255}
	//p.Add(l)

	movingAverage := calcs.GetMovingAverage(pts, numPoints)
	m, err := plotter.NewLine(movingAverage)
	if err != nil {
		panic(err)
	}
	m.LineStyle.Width = vg.Points(1)
	m.LineStyle.Color = color.RGBA{R: 200, A: 255}
	p.Add(m)

	standardDeviation := calcs.GetStandardDeviation(pts, movingAverage, numPoints)

	upperBollingerBand := plotter.XYs{}
	lowerBollingerBand := plotter.XYs{}
	for i, sd := range standardDeviation {
		upperBollingerBand = append(upperBollingerBand, plotter.XY{X: sd.X, Y: movingAverage[i].Y + (2 * sd.Y)})
		lowerBollingerBand = append(lowerBollingerBand, plotter.XY{X: sd.X, Y: movingAverage[i].Y - (2 * sd.Y)})
	}
	ub, err := plotter.NewLine(upperBollingerBand)
	if err != nil {
		panic(err)
	}
	ub.LineStyle.Width = vg.Points(1)
	ub.LineStyle.Color = color.RGBA{G: 200, A: 255}
	p.Add(ub)

	lb, err := plotter.NewLine(lowerBollingerBand)
	if err != nil {
		panic(err)
	}
	lb.LineStyle.Width = vg.Points(1)
	lb.LineStyle.Color = color.RGBA{B: 200, A: 255}
	p.Add(lb)

	// Add candlesticks
	bars, err := custplotter.NewCandlesticks(hlcvs)
	if err != nil {
		panic(err)
	}
	p.Add(bars)

	// Save the plot to a PNG file.
	if err := p.Save(64*vg.Inch, 16*vg.Inch, "points.png"); err != nil {
		panic(err)
	}
}
