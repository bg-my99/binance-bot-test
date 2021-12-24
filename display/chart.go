package display

import (
	"image/color"
	"os"

	"github.com/pplcc/plotext/custplotter"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/font"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgimg"

	candles "binance-bot-test/storage"
)

type ChartLine struct {
	width         int
	color         color.RGBA
	PointsChannel chan plotter.XY
	points        plotter.XYs
}

type ChartBar struct {
	color         color.RGBA
	PointsChannel chan float64
	points        plotter.Values
}

type ChartCandles struct {
	CandlesChannel chan *candles.Candle
	candles        []*candles.Candle
}

type Chart struct {
	plot    *plot.Plot
	lines   []*ChartLine
	bars    []*ChartBar
	candles *ChartCandles
	rsi     *ChartLine
}

func (c *Chart) Init() {
	c.plot = plot.New()
	c.plot.Title.Text = "Chart example"
	c.plot.X.Label.Text = "X"
	c.plot.Y.Label.Text = "Y"

	xticks := plot.TimeTicks{Format: "2006-01-02\n15:04"}
	c.plot.X.Tick.Marker = xticks
}

func (c *Chart) AddLine(r uint8, g uint8, b uint8, width int) *ChartLine {
	newChartLine := &ChartLine{width: width, color: color.RGBA{R: r, G: g, B: b, A: 255}}
	c.lines = append(c.lines, newChartLine)

	newChartLine.PointsChannel = make(chan plotter.XY)

	go func(cl *ChartLine) {
		for point := range newChartLine.PointsChannel {
			cl.points = append(newChartLine.points, point)
		}
	}(newChartLine)
	return newChartLine
}

func (c *Chart) AddRsi() *ChartLine {
	newChartLine := &ChartLine{width: 2, color: color.RGBA{R: 100, G: 200, B: 250, A: 255}}
	c.rsi = newChartLine

	newChartLine.PointsChannel = make(chan plotter.XY)

	go func(cl *ChartLine) {
		for point := range newChartLine.PointsChannel {
			cl.points = append(newChartLine.points, point)
		}
	}(newChartLine)
	return newChartLine
}

func (c *Chart) AddCandles() *ChartCandles {
	c.candles = &ChartCandles{}

	c.candles.CandlesChannel = make(chan *candles.Candle)

	go func(cc *ChartCandles) {
		for candle := range cc.CandlesChannel {
			cc.candles = append(cc.candles, candle)
		}
	}(c.candles)
	return c.candles
}

func (c *Chart) AddBar(r uint8, g uint8, b uint8) *ChartBar {
	newChartBar := &ChartBar{color: color.RGBA{R: r, G: g, B: b, A: 255}}

	c.bars = append(c.bars, newChartBar)

	newChartBar.PointsChannel = make(chan float64)

	go func(cb *ChartBar) {
		for point := range cb.PointsChannel {
			cb.points = append(cb.points, point)
		}
	}(newChartBar)
	return newChartBar
}

func (c *Chart) Output() {
	// Make sure all the channels are done then add the lines to the plot
	if c.candles != nil {
		// Add candlesticks
		candles, err := custplotter.NewCandlesticks(candles.GetHLCVs(c.candles.candles))
		if err != nil {
			panic(err)
		}
		c.plot.Add(candles)
	}
	for _, line := range c.lines {
		close(line.PointsChannel)

		m, err := plotter.NewLine(line.points)
		if err != nil {
			panic(err)
		}
		m.LineStyle.Width = vg.Points(float64(line.width))
		m.LineStyle.Color = line.color
		c.plot.Add(m)
	}
	for _, barchart := range c.bars {
		close(barchart.PointsChannel)

		barChart, err := plotter.NewBarChart(barchart.points, font.Length(192*float64(vg.Inch)/float64(barchart.points.Len())))
		if err != nil {
			panic(err)
		}
		c.plot.Add(barChart)
	}
	if c.rsi != nil {

		w, err := os.Create("aligned.png")
		if err != nil {
			panic(err)
		}

		img := vgimg.New(192*vg.Inch, 32*vg.Inch)
		dc := draw.New(img)

		plots := make([][]*plot.Plot, 2)
		plots[0] = make([]*plot.Plot, 1)
		plots[0][0] = c.plot

		plt := plot.New()
		plt.Title.Text = "RSI"
		plt.X.Label.Text = "X"
		plt.Y.Label.Text = "Y"

		xticks := plot.TimeTicks{Format: "2006-01-02\n15:04"}
		plt.X.Tick.Marker = xticks

		close(c.rsi.PointsChannel)

		m, err := plotter.NewLine(c.rsi.points)
		if err != nil {
			panic(err)
		}
		m.LineStyle.Width = vg.Points(float64(c.rsi.width))
		m.LineStyle.Color = c.rsi.color
		plt.Add(m)

		plots[1] = make([]*plot.Plot, 1)
		plots[1][0] = plt

		t := draw.Tiles{
			Rows: 2,
			Cols: 1,
		}
		canvases := plot.Align(plots, t, dc)
		plots[0][0].Draw(canvases[0][0])
		plots[1][0].Draw(canvases[1][0])

		png := vgimg.PngCanvas{Canvas: img}
		if _, err := png.WriteTo(w); err != nil {
			panic(err)
		}
	} else {
		// Save the plot to a PNG file.
		if err := c.plot.Save(192*vg.Inch, 48*vg.Inch, "fractalpoints.png"); err != nil {
			panic(err)
		}
	}
}
