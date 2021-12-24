// json.go
package main

import (
	"net/http"
)

type ValueLine struct {
	X []float64 `json:"x"`
	Y []float64 `json:"y"`
	Z []float64 `json:"z"`
}

type ValueLines struct {
	Lines []ValueLine `json:"lines"`
}

func main() {

	http.HandleFunc("/encode", func(w http.ResponseWriter, r *http.Request) {
		/*values := ValueLines{}
		xRange := 20
		yRange := 20
		values.Lines = make([]ValueLine, yRange)
		for y := 0; y < yRange; y++ {

			values.Lines[y].X = make([]float64, xRange)
			values.Lines[y].Y = make([]float64, xRange)
			values.Lines[y].Z = make([]float64, xRange)
			for x := 0; x < xRange; x++ {
				values.Lines[y].X[x] = float64(x)
				values.Lines[y].Y[x] = float64(y)
				values.Lines[y].Z[x] = rand.Float64() * 2.0
			}
		}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(values)*/
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.ServeFile(w, r, "/home/chris/Projects/crypto/binance-bot-test/results.json")
	})

	http.ListenAndServe(":8080", nil)
}
