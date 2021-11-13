package calcs

import (
	"math"

	"gonum.org/v1/plot/plotter"
)

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
