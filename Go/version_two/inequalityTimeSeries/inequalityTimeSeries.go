package inequalityTimeSeries

import (
	"encoding/csv"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"io"
	"log"
	"log/slog"
	"os"
	"slices"
	"strconv"
)

type RegionAndPercentile struct {
	Region     string
	Percentile string
}

type TimeSeriesData struct {
	CSVFileColumnToRegionNameMap map[int]string
	ColumnRenamingMap            map[string]string
	Years                        []int
	ExistingRAndPs               []RegionAndPercentile
	TimeSeries                   map[RegionAndPercentile][]opts.LineData
}

func (ts *TimeSeriesData) SetCSVFileColumnToRegionNameMap(mapToUse map[int]string) {
	ts.CSVFileColumnToRegionNameMap = mapToUse
}

func (ts *TimeSeriesData) SetColumnRenamingmap(mapToUse map[string]string) {
	ts.ColumnRenamingMap = mapToUse
}

// addDataPoint assumes that values are integers. I've used integers here to make the data points
// on the line produced more visually consistent.
// Please note that this introduces a larger rounding error than using float64s.
func (ts *TimeSeriesData) addDataPoint(rAndP RegionAndPercentile, value int, setToNil bool) {
	if !slices.Contains(ts.ExistingRAndPs, rAndP) {
		ts.TimeSeries[rAndP] = []opts.LineData{}
		ts.ExistingRAndPs = append(ts.ExistingRAndPs, rAndP)
	}

	if setToNil {
		ts.TimeSeries[rAndP] = append(ts.TimeSeries[rAndP], opts.LineData{Value: nil})
	} else {
		ts.TimeSeries[rAndP] = append(ts.TimeSeries[rAndP], opts.LineData{Value: value})
	}
}

func (ts *TimeSeriesData) appendIfNotInYearsList(newYear int) {
	if !slices.Contains(ts.Years, newYear) {
		ts.Years = append(ts.Years, newYear)
	}
}

func (ts *TimeSeriesData) LoadCSVDataIntoInequalityTimeSeries(pathToCSV string) {
	file, err := os.Open(pathToCSV)
	if err != nil {
		log.Print("error when reading file:", err)
	}
	defer file.Close()

	csvR := csv.NewReader(file)
	csvR.Comma = ';'

	// skip first row, which contains the column names
	csvR.Read()

	for {
		newLine, err := csvR.Read()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Println("error when reading new line from file:", err)
		}

		ts.updateYearsAndRegionAndPercentileDataWithNewRow(newLine)
	}
}

func (ts *TimeSeriesData) updateYearsAndRegionAndPercentileDataWithNewRow(newLine []string) {
	yearVal, err := strconv.Atoi(newLine[1])
	if err != nil {
		slog.Error(
			"could not convert string of year value to int",
			"val in:", newLine[1],
			"err", err,
		)
	}
	ts.appendIfNotInYearsList(yearVal)

	for i := 2; i < len(newLine); i++ {
		if newLine[i] == "" {
			ts.addDataPoint(
				RegionAndPercentile{Region: ts.CSVFileColumnToRegionNameMap[i], Percentile: newLine[0]},
				0,
				true,
			)

		} else {
			valInCol, err := strconv.ParseFloat(newLine[i], 64)
			if err != nil {
				slog.Error(
					"unable to parse string into float",
					"string:", newLine[i],
					"err", err,
				)
			}
			ts.addDataPoint(
				RegionAndPercentile{Region: ts.CSVFileColumnToRegionNameMap[i], Percentile: newLine[0]},
				int(valInCol*100),
				false,
			)
		}
	}
}

func countDigitsInInt(intIn int) int {
	numDigits := 0
	for intIn > 0 {
		numDigits++
		intIn /= 10
	}
	return numDigits
}

func (ts *TimeSeriesData) CreateAndDisplayTimeSeriesLineChart(w io.Writer) {
	linesChart := charts.NewLine()

	linesChart.SetXAxis(ts.Years)

	for key, timeSeries := range ts.TimeSeries {
		linesChart.AddSeries(
			key.Region+" "+ts.ColumnRenamingMap[key.Percentile],
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

	linesChart.Render(w)
}
