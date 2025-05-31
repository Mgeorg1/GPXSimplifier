package main

import (
	"encoding/csv"
	"encoding/xml"
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"
)

const EarthRadius = 6371000 // in meters

func toRadians(deg float64) float64 {
	return deg * math.Pi / 180
}

func haversineWithAltitude(lat1, lon1, alt1, lat2, lon2, alt2 float64) float64 {
	horizontalDist := haversine(lat1, lon1, lat2, lon2)
	verticalDist := alt2 - alt1
	return math.Sqrt(horizontalDist*horizontalDist + verticalDist*verticalDist)
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	lat1 = toRadians(lat1)
	lat2 = toRadians(lat2)
	lon1 = toRadians(lon1)
	lon2 = toRadians(lon2)
	dLat := lat2 - lat1
	dLon := lon2 - lon1

	sinDLat := math.Sin(dLat / 2)
	sinDLon := math.Sin(dLon / 2)

	h := sinDLat*sinDLat + math.Cos(lat1)*math.Cos(lat2)*sinDLon*sinDLon
	c := 2 * math.Asin(math.Sqrt(h))
	return EarthRadius * c
}

type GPX struct {
	XMLName  xml.Name  `xml:"gpx"`
	Segments []Segment `xml:"trk>trkseg"`
}

type Segment struct {
	TrkPts []TrackPoint `xml:"trkpt"`
}

type TrackPoint struct {
	Lat  float64   `xml:"lat,attr"`
	Lon  float64   `xml:"lon,attr"`
	Ele  float64   `xml:"ele"`
	Time time.Time `xml:"time"`
	HR   int       `xml:"extensions>TrackPointExtension>hr"`
}

type Extension struct {
	TPX TrackPointExtension `xml:"TrackPointExtension"`
}

type TrackPointExtension struct {
	HR int `xml:"hr"`
}

func main() {
	var (
		outputFilePath string
		saveIntervalM  float64
		inputFilePath  string
	)
	flag.StringVar(&outputFilePath, "out", "output.csv", "Output CSV file path (optional)")
	flag.Float64Var(&saveIntervalM, "interval", 200, "Distance interval in meters for saving points (optional)")
	flag.StringVar(&inputFilePath, "input", "", "Input GPX file path (required)")
	flag.Parse()

	if inputFilePath == "" {
		fmt.Println("Usage: GPXSimplifier -input <input.gpx> [-out output.csv] [-interval meters]")
		return
	}

	gpxFile, err := os.Open(inputFilePath)
	if err != nil {
		println("Error opening file:", err.Error())
		return
	}
	defer gpxFile.Close()
	decoder := xml.NewDecoder(gpxFile)
	var gpx GPX
	err = decoder.Decode(&gpx)
	if err != nil {
		println("Error decoding XML:", err.Error())
		return
	}
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		println("Error creating output file:", err.Error())
		return
	}
	defer outputFile.Close()
	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	writer.Write([]string{"distance_m", "timestamp", "ele", "hr", "pace_min_per_km"})

	var totalDist float64
	var lastPoint *TrackPoint
	var lastDist float64
	var lastTime time.Time

	for _, seg := range gpx.Segments {
		for i, pt := range seg.TrkPts {
			if lastPoint == nil {
				lastPoint = &seg.TrkPts[i]
				lastTime = pt.Time
				continue
			}

			dist := haversineWithAltitude(lastPoint.Lat, lastPoint.Lon, lastPoint.Ele, pt.Lat, pt.Lon, pt.Ele)
			totalDist += dist
			if totalDist-lastDist >= saveIntervalM {
				timeDiff := pt.Time.Sub(lastTime).Minutes()
				if timeDiff == 0 {
					timeDiff = 0.0001
				}
				pace := (totalDist - lastDist) / (timeDiff * 1000) // km/min
				paceMinPerKm := 1 / pace                           // min/km
				writer.Write([]string{
					fmt.Sprintf("%.1f", totalDist),
					pt.Time.Format(time.RFC3339),
					fmt.Sprintf("%.1f", pt.Ele),
					strconv.Itoa(pt.HR),
					fmt.Sprintf("%.2f", paceMinPerKm),
				})
				lastDist = totalDist
				lastTime = pt.Time
			}
			lastPoint = &seg.TrkPts[i]
		}
	}
	fmt.Println("Output written to ", outputFilePath)
}
