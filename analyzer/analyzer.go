package analyzer

import (
	"encoding/xml"
	"math"
	"os"
	"time"

	"github.com/tkrajina/gpxgo/gpx"
)

type FallEvent struct {
	Timestamp time.Time
	Latitude  float64
	Longitude float64
}

// AnalyzeGPXFile analyzes a GPX file for potential falls
func AnalyzeGPXFile(filename string) ([]FallEvent, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var gpxDoc gpx.GPX
	if err := xml.Unmarshal(data, &gpxDoc); err != nil {
		return nil, err
	}

	return analyzeTracks(gpxDoc.Tracks), nil
}

func analyzeTracks(tracks []gpx.GPXTrack) []FallEvent {
	var falls []FallEvent
	
	for _, track := range tracks {
		for _, segment := range track.Segments {
			falls = append(falls, analyzeSegment(segment)...)
		}
	}
	
	return falls
}

func analyzeSegment(segment gpx.GPXTrackSegment) []FallEvent {
	var falls []FallEvent
	
	// Need at least 3 points to detect a fall
	if len(segment.Points) < 3 {
		return falls
	}

	// Parameters for fall detection
	const (
		timeWindow          = 5     // seconds
		elevationChange     = -0.5  // meters
		maxHorizontalSpeed  = 0.3   // meters per second - max allowed horizontal movement
		minVerticalSpeed    = 0.1   // meters per second - minimum vertical speed
	)

	for i := 1; i < len(segment.Points)-1; i++ {
		prev := segment.Points[i-1]
		curr := segment.Points[i]
		next := segment.Points[i+1]

		// Calculate time difference
		timeDiff := next.Timestamp.Sub(prev.Timestamp).Seconds()

		// Skip if time difference is too large
		if timeDiff > timeWindow {
			continue
		}

		// Check if elevations are present and valid
		if !next.Elevation.NotNull() || !prev.Elevation.NotNull() {
			continue
		}

		// Calculate horizontal speed
		horizontalSpeed := calculateHorizontalSpeed(prev, next)
		
		// Calculate vertical speed (negative means dropping)
		verticalSpeed := (next.Elevation.Value() - prev.Elevation.Value()) / timeDiff
		
		// Check for vertical drop with minimal horizontal movement
		if horizontalSpeed <= maxHorizontalSpeed &&
			verticalSpeed < -minVerticalSpeed &&
			(next.Elevation.Value()-prev.Elevation.Value()) < elevationChange {

			falls = append(falls, FallEvent{
				Timestamp: curr.Timestamp,
				Latitude:  curr.Latitude,
				Longitude: curr.Longitude,
			})
		}
	}

	return falls
}

func calculateSpeed(p1, p2 gpx.GPXPoint) float64 {
	distance := calculateDistance(p1, p2)
	timeDiff := p2.Timestamp.Sub(p1.Timestamp).Seconds()
	if timeDiff == 0 {
		return 0
	}
	return distance / timeDiff
}

func calculateDistance(p1, p2 gpx.GPXPoint) float64 {
	const earthRadius = 6371000 // meters

	lat1 := toRadians(p1.Latitude)
	lon1 := toRadians(p1.Longitude)
	lat2 := toRadians(p2.Latitude)
	lon2 := toRadians(p2.Longitude)

	dLat := lat2 - lat1
	dLon := lon2 - lon1

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func calculateHorizontalSpeed(p1, p2 gpx.GPXPoint) float64 {
	horizontalDistance := calculateDistance(p1, p2)
	timeDiff := p2.Timestamp.Sub(p1.Timestamp).Seconds()
	if timeDiff == 0 {
		return 0
	}
	return horizontalDistance / timeDiff
}
