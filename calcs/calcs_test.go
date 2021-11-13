package calcs

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
	"testing"
	"time"

	"gonum.org/v1/plot/plotter"
)

func TestGetMovingAverage(t *testing.T) {

	filename := "test-calcs.1.csv"
	f, err := os.Open(filename)
	if err != nil {
		t.Fatal("Unable to read input file "+filename, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		t.Fatal("Unable to parse file as CSV for "+filename, err)
	}

	pts := plotter.XYs{}
	expectedMA := plotter.XYs{}
	for i, line := range records {
		timestamp, err := time.Parse("02/01/2006", line[0])
		if err != nil {
			fmt.Println(err)
			continue
		}
		value, err := strconv.ParseFloat(line[1], 64)
		if err != nil {
			continue
		}
		ma, err := strconv.ParseFloat(line[2], 64)
		if err != nil {
			continue
		}
		pts = append(pts, plotter.XY{X: float64(timestamp.UnixNano()), Y: value})
		if i >= 19 {
			expectedMA = append(expectedMA, plotter.XY{X: float64(timestamp.UnixNano()), Y: ma})
		}
	}
	ma := GetMovingAverage(pts, 20)

	if len(ma) != len(expectedMA) {
		t.Errorf("len(ma)(%d) doesnt match expected(%d)\n", len(ma), len(expectedMA))
	}
	tolerance := 0.00000001
	for i, expected := range expectedMA {
		if !(math.Abs(expected.Y-ma[i].Y) < tolerance) {
			fmt.Printf("GetMovingAverage failed at %d:\n%v\n", i, expected.Y)
			fmt.Printf("vs:\n%v\n", ma[i].Y)
			t.Errorf("GetMovingAverage failed")
		}
	}
}

func TestGetStandardDeviation(t *testing.T) {

	filename := "test-calcs.1.csv"
	f, err := os.Open(filename)
	if err != nil {
		t.Fatal("Unable to read input file "+filename, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		t.Fatal("Unable to parse file as CSV for "+filename, err)
	}

	pts := plotter.XYs{}
	movingAverage := plotter.XYs{}
	expectedStdDev := plotter.XYs{}
	for i, line := range records {
		timestamp, err := time.Parse("02/01/2006", line[0])
		if err != nil {
			fmt.Println(err)
			continue
		}
		value, err := strconv.ParseFloat(line[1], 64)
		if err != nil {
			continue
		}
		ma, err := strconv.ParseFloat(line[2], 64)
		if err != nil {
			continue
		}
		bol, err := strconv.ParseFloat(line[3], 64)
		if err != nil {
			continue
		}
		ts := float64(timestamp.UnixNano())
		pts = append(pts, plotter.XY{X: ts, Y: value})
		if i >= 19 {
			movingAverage = append(movingAverage, plotter.XY{X: ts, Y: ma})
			expectedStdDev = append(expectedStdDev, plotter.XY{X: ts, Y: (bol - ma) / 2.0})
		}
	}
	sd := GetStandardDeviation(pts, movingAverage, 20)

	if len(sd) != len(expectedStdDev) {
		t.Errorf("len(ma)(%d) doesnt match expected(%d)\n", len(sd), len(expectedStdDev))
	}
	tolerance := 0.00000001
	for i, expected := range expectedStdDev {
		if !(math.Abs(expected.Y-expectedStdDev[i].Y) < tolerance) {
			fmt.Printf("GetStandardDeviation failed at %d:\n%v\n", i, expected.Y)
			fmt.Printf("vs:\n%v\n", sd[i].Y)
			t.Errorf("GetStandardDeviation failed")
		}
	}
}
