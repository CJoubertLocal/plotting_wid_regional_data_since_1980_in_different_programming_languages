package main

import (
	"github.com/CJoubertLocal/plotting_wid_regional_data_with_go.git/inequalityTimeSeries"
	"github.com/go-echarts/go-echarts/v2/opts"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", displayInequalityDataChart)
	http.ListenAndServe(":8080", nil)
}

func displayInequalityDataChart(w http.ResponseWriter, _ *http.Request) {
	var regionAndPercentileData = inequalityTimeSeries.TimeSeriesData{
		Years:          []int{},
		ExistingRAndPs: []inequalityTimeSeries.RegionAndPercentile{},
		TimeSeries:     map[inequalityTimeSeries.RegionAndPercentile][]opts.LineData{},
	}
	regionAndPercentileData.SetCSVFileColumnToRegionNameMap(
		map[int]string{
			2: "Africa",
			3: "Asia",
			4: "Latin America",
			5: "Europe",
			6: "MiddleEast",
			7: "Oceania",
			8: "North America",
		},
	)
	regionAndPercentileData.SetColumnRenamingmap(
		map[string]string{
			"p0p50":   "Bottom 50%",
			"p90p100": "Top 10%",
			"p99p100": "Top 1%",
		},
	)
	regionAndPercentileData.LoadCSVDataIntoInequalityTimeSeries(os.Args[1])
	regionAndPercentileData.CreateAndDisplayTimeSeriesLineChart(w)
}
