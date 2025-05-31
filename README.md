# GPXSimplifier
GPX Simplifier is a Go tool to parse GPX files, calculate distances considering elevation changes, and export simplified track points at configurable distance intervals into CSV format. It also computes pace (min/km) between saved points.
Primarily designed to preprocess and simplify GPX data for uploading into GPT-based training assistant.
Main 
## Features

* Parses GPX tracks with latitude, longitude, elevation, timestamp, and heart rate.

* Calculates accurate distances using horizontal distance plus elevation differences.

* Saves points every configurable distance interval (default 200 meters).

* Outputs CSV with distance, time, elevation, heart rate, and pace.

## Usage

`go run main.go -input input.gpx [-out output.csv] [-interval 200]`
