package entergo
import (
	"github.com/artonge/go-csv-tag/v2"
	"github.com/pion/dtls/v2"
)
type TimeSeriesLightsCsv struct {
	Time int  `csv:"Time"`
	ID1   int     `csv:"ID1"`
	Red1 int  `csv:"Red1"`
	Blue1 int  `csv:"Blue1"`
	Green1 int  `csv:"Green1"`
	ID2   int     `csv:"ID2"`
	Red2 int  `csv:"Red2"`
	Blue2 int  `csv:"Blue2"`
	Green2 int  `csv:"Green2"`
}

type TimeSeriesLights struct {
	LightPackets []LightPacket
}

func CreateDBFromTimeseries(timeseries []TimeSeriesLightsCsv) {
	err := csvtag.DumpToFile(timeseries, "timeseries.csv")
	if err != nil {
		print("Error writing file")
	}
}


func ReadFromCsv()(TimeSeries []TimeSeriesLights, Interval int) {
	tab := []TimeSeriesLightsCsv{}                                   // Create the slice where to put the content
	err  := csvtag.LoadFromPath(
		"timeseries.csv",                                   // Path of the csv file
		&tab,                                         // A pointer to the create slice
		csvtag.CsvOptions{                            // Load your csv with optional options
			Separator: ',',                           // changes the values separator, default to ',' 
	})
	if err != nil {
		print("Error reading file")
	}
	timeseries, interval := convtimeseries(tab)
	return timeseries, interval
}



func convtimeseries(timeseries []TimeSeriesLightsCsv)(TimeSeries []TimeSeriesLights, Interval int){
	output := []TimeSeriesLights{}
	for i, e := range(timeseries) {
		if i == 0 {
			Interval = e.ID1
		} else {
			packets := []LightPacket{LightPacket{LightID: uint8(e.ID1), Red: uint8(e.Red1), Blue: uint8(e.Blue1), Green: uint8(e.Green1)}}
			packets = append(packets[:], LightPacket{LightID: uint8(e.ID1), Red: uint8(e.Red1), Blue: uint8(e.Blue1), Green: uint8(e.Green1)})
			output = append(output, TimeSeriesLights{packets[:]})
		}
	}
	return output, Interval
}

func PlayFromTimeSeries (timeseries []TimeSeriesLights, interval int, connection *dtls.Conn, status chan []LightPacket) {
	for _, e := range(timeseries) {
		status <- e.LightPackets
	}
	defer StreamLoop(interval, connection, status)
}
