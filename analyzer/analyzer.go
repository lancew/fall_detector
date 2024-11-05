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
		speedThreshold    = 0.5  // meters per second
		timeWindow       = 5    // seconds
		elevationChange  = -2.0 // meters
	)

	for i := 1; i < len(segment.Points)-1; i++ {
		prev := segment.Points[i-1]
		curr := segment.Points[i]
		next := segment.Points[i+1]

		// Calculate speed changes
		speedBefore := calculateSpeed(prev, curr)
		speedAfter := calculateSpeed(curr, next)

		// Calculate time differences
		timeDiff := next.Timestamp.Sub(prev.Timestamp).Seconds()

		// Check for sudden speed decrease and elevation drop within time window
		if timeDiff <= timeWindow &&
			speedBefore > speedThreshold &&
			speedAfter < speedThreshold &&
			next.Elevation.Value() != 0 && prev.Elevation.Value() != 0 &&
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
