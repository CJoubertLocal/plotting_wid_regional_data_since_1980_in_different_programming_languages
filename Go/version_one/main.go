package main

import (
	"encoding/csv"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"slices"
	"sort"
	"strconv"
)

// note: this is very rigid. Here, I am assuming that the file which comes in has
// this exact number of columns.
type rowOfRegionInequalityData struct {
	Percentile   string
	Year         int
	Africa       float64
	Asia         float64
	LatinAmerica float64
	Europe       float64
	MiddleEast   float64
	Oceania      float64
	NorthAmerica float64
}

type regionAndPercentile struct {
	Region     string
	Percentile string
}

func main() {
	http.HandleFunc("/", displayInequalityDataChart)
	http.ListenAndServe(":8080", nil)
}

func displayInequalityDataChart(w http.ResponseWriter, _ *http.Request) {
	rowsOfData := loadCSVData(os.Args[1])
	timeSeriesData, years := convertRowsOfInequalityDataIntoMapOfRegionPercentileToLineData(rowsOfData)
	linesChart := createTimeSeriesLineChart(timeSeriesData, years)
	linesChart.Render(w)
}

func loadCSVData(pathToCSV string) []rowOfRegionInequalityData {
	// An empty slice is declared here so that there is no nil pointer exception
	// when appending to this slice later on.
	var csvData = []rowOfRegionInequalityData{}

	file, err := os.Open(pathToCSV)
	if err != nil {
		log.Print("error when reading file:", err)
	}
	defer file.Close()

	csvR := csv.NewReader(file)
	csvR.Comma = ';'

	// skip first row
	csvR.Read()

	for {
		newLine, err := csvR.Read()
		if err == io.EOF {
			return csvData
		}
		if err != nil {
			log.Println("error when reading new line from file:", err)
		}

		csvData = append(csvData, convertSliceOfStingDataIntoRowOfRegionInequalityData(newLine))
	}

	return csvData
}

// convertSliceOfStingDataIntoRowOfRegionInequalityData assumes that rowIn has exactly
// the number of strings as are required by rowOfRegionInequalityData and that
// those strings are in the same order as the fields in rowOfRegionInequalityData.
// If an inequality value for a region and percentile cannot be converted into a float
// then I assume it is missing and return '-1', which should be an impossible value.
// This needs to be filtered out by later functions.
func convertSliceOfStingDataIntoRowOfRegionInequalityData(rowIn []string) rowOfRegionInequalityData {
	var rowToReturn = rowOfRegionInequalityData{
		Percentile: rowIn[0],
	}

	yearInt, err := strconv.Atoi(rowIn[1])
	if err != nil {
		log.Println("unable to convert string of year value into int:", err)
	}
	rowToReturn.Year = yearInt

	if rowIn[2] == "" {
		rowToReturn.Africa = -1
	} else {
		rowToReturn.Africa, err = strconv.ParseFloat(rowIn[2], 32)
		if err != nil {
			log.Println("unable to convert string for Africa value into int:", err)
		}
	}

	if rowIn[3] == "" {
		rowToReturn.Asia = -1
	} else {
		rowToReturn.Asia, err = strconv.ParseFloat(rowIn[3], 32)
		if err != nil {
			log.Println("unable to convert string for Asia value into int:", err)
		}
	}

	if rowIn[4] == "" {
		rowToReturn.LatinAmerica = -1
	} else {
		rowToReturn.LatinAmerica, err = strconv.ParseFloat(rowIn[4], 32)
		if err != nil {
			log.Println("unable to convert string for Latin America value into int:", err)
		}
	}

	if rowIn[5] == "" {
		rowToReturn.Europe = -1
	} else {
		rowToReturn.Europe, err = strconv.ParseFloat(rowIn[5], 32)
		if err != nil {
			log.Println("unable to convert string for Europe value into int:", err)
		}
	}

	if rowIn[6] == "" {
		rowToReturn.MiddleEast = -1
	} else {
		rowToReturn.MiddleEast, err = strconv.ParseFloat(rowIn[6], 32)
		if err != nil {
			log.Println("unable to convert string for Middle East value into int:", err)
		}
	}

	if rowIn[7] == "" {
		rowToReturn.Oceania = -1
	} else {
		rowToReturn.Oceania, err = strconv.ParseFloat(rowIn[7], 32)
		if err != nil {
			log.Println("unable to convert string for Oceania value into int:", err)
		}
	}

	if rowIn[8] == "" {
		rowToReturn.NorthAmerica = -1
	} else {
		rowToReturn.NorthAmerica, err = strconv.ParseFloat(rowIn[8], 32)
		if err != nil {
			log.Println("unable to convert string for North America value into int:", err)
		}
	}

	return rowToReturn
}

// sortRowsOfInequalityByYear sorts values in place.
func sortRowsOfInequalityByYear(ineqRows []rowOfRegionInequalityData) {
	sort.Slice(ineqRows, func(i int, j int) bool {
		return ineqRows[i].Year < ineqRows[j].Year
	})
}

// convertRowsOfInequalityDataIntoMapOfRegionPercentileToLineData assumes that
// the header row is not included in the ineqRows data.
// It further assumes that ineqRows are sorted by Year.
// If a 'value' is -1, then it is assumed to be a missing value.
// In this case, nil will be included in the opts result, so that there are
// enough opts.LineData to plot on the graph, but missing variables are not
// shown.
func convertRowsOfInequalityDataIntoMapOfRegionPercentileToLineData(ineqRows []rowOfRegionInequalityData) (map[regionAndPercentile][]opts.LineData, []int) {
	var regionAndPercentileToData = map[regionAndPercentile][]opts.LineData{}
	var years = []int{}
	var existingRandPs = []regionAndPercentile{}

	for _, r := range ineqRows {
		if !slices.Contains(years, r.Year) {
			years = append(years, r.Year)
		}

		// reflection is used here to avoid duplicating code.
		fields := reflect.ValueOf(r)
		for i := 2; i < fields.NumField(); i++ {
			fieldName := fields.Type().Field(i).Name
			fieldValue := fields.Field(i)

			rAndP := regionAndPercentile{Region: fieldName, Percentile: r.Percentile}
			if !slices.Contains(existingRandPs, rAndP) {
				regionAndPercentileToData[rAndP] = []opts.LineData{}
				existingRandPs = append(existingRandPs, rAndP)
			}

			if fieldValue.Float() != -1.0 {
				regionAndPercentileToData[rAndP] = append(regionAndPercentileToData[rAndP], opts.LineData{Value: fieldValue.Float() * 100})
			} else {
				regionAndPercentileToData[rAndP] = append(regionAndPercentileToData[rAndP], opts.LineData{Value: nil})
			}

		}
	}

	return regionAndPercentileToData, years
}

var percentileMap = map[string]string{
	"p0p50":   "Bottom 50%",
	"p90p100": "Top 10%",
	"p99p100": "Top 1%",
}

func createTimeSeriesLineChart(timeSeriesData map[regionAndPercentile][]opts.LineData, years []int) *charts.Line {
	linesChart := charts.NewLine()

	linesChart.SetXAxis(years)

	for key, timeSeries := range timeSeriesData {
		linesChart.AddSeries(
			key.Region+" "+percentileMap[key.Percentile],
			timeSeries,
		).SetSeriesOptions(charts.WithLabelOpts(
			opts.Label{
				Show:     true,
				Position: "top",
			}))
	}

	linesChart.SetGlobalOptions(
		charts.WithInitializationOpts(
			opts.Initialization{
				PageTitle: "Inequality plot",
			},
		),
		charts.WithTitleOpts(
			opts.Title{
				Title: "Estimated Proportions of National Income Owned by Different Population Percentiles",
				Left:  "center",
			},
		),
		charts.WithXAxisOpts(
			opts.XAxis{
				Show: true,
				Name: "Year",
			},
		),
		charts.WithYAxisOpts(
			opts.YAxis{
				Show:         true,
				Name:         "Proportion of National Income (%)",
				NameLocation: "middle",
			},
		),
		charts.WithLegendOpts(
			opts.Legend{
				Show:      true,
				ItemWidth: 30,
				Type:      "scroll",
				Top:       "30",
			},
		),
		charts.WithDataZoomOpts(
			opts.DataZoom{
				Start:      60,
				End:        100,
				XAxisIndex: []int{0},
			},
		),
		charts.WithTooltipOpts(
			opts.Tooltip{
				Show:    true,
				Trigger: "axis",
			},
		),
	)

	return linesChart
}
