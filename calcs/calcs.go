package calcs

import (
	"math"

	"gonum.org/v1/plot/plotter"
)

func CalcExpMovingAverages(price float64, currentEMA20 float64, currentEMA50 float64, currentEMA100 float64) (float64, float64, float64) {
	eMA20Multiplier := (2.0 / (20.0 + 1.0))
	eMA50Multiplier := (2.0 / (50.0 + 1.0))
	eMA100Multiplier := (2.0 / (100.0 + 1.0))

	newEMA20 := (price * eMA20Multiplier) + (currentEMA20 * (1 - eMA20Multiplier))
	newEMA50 := (price * eMA50Multiplier) + (currentEMA50 * (1 - eMA50Multiplier))
	newEMA100 := (price * eMA100Multiplier) + (currentEMA100 * (1 - eMA100Multiplier))

	return newEMA20, newEMA50, newEMA100
}

func GetExpMovingAverages(pts plotter.XYs, numPoints int) (plotter.XYs, plotter.XYs, plotter.XYs) {
	eMA20 := plotter.XYs{}
	eMA50 := plotter.XYs{}
	eMA100 := plotter.XYs{}

	runningEMA20 := pts[0].Y
	runningEMA50 := pts[0].Y
	runningEMA100 := pts[0].Y
	for _, trade := range pts {

		runningEMA20, runningEMA50, runningEMA100 = CalcExpMovingAverages(trade.Y, runningEMA20, runningEMA50, runningEMA100)

		eMA20 = append(eMA20, plotter.XY{X: trade.X, Y: runningEMA20})
		eMA50 = append(eMA50, plotter.XY{X: trade.X, Y: runningEMA50})
		eMA100 = append(eMA100, plotter.XY{X: trade.X, Y: runningEMA100})
	}
	return eMA20, eMA50, eMA100
}

func GetMovingAverage(pts plotter.XYs, numPoints int) plotter.XYs {
	movingAverage := plotter.XYs{}
	runningCount := 0.0
	for i, trade := range pts {
		runningCount += trade.Y
		if i >= (numPoints - 1) {
			movingAverage = append(movingAverage, plotter.XY{X: trade.X, Y: runningCount / float64(numPoints)})
			runningCount -= pts[i-(numPoints-1)].Y
		}
	}
	return movingAverage
}

func GetStandardDeviation(pts plotter.XYs, movingAverage plotter.XYs, numPoints int) plotter.XYs {
	standardDeviation := plotter.XYs{}
	for j, ma := range movingAverage {
		runningCount := 0.0
		for i := j; i < (j + numPoints); i++ {
			runningCount += math.Pow(pts[i].Y-ma.Y, 2)
		}
		standardDeviation = append(standardDeviation, plotter.XY{X: ma.X, Y: math.Sqrt(runningCount / float64(numPoints-1))})
	}
	return standardDeviation
}
