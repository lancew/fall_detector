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
	Severity  string // Added to indicate fall severity
}

// Constants for fall detection - adjusted to be extremely sensitive
const (
	timeWindow          = 1.0  // seconds (even shorter time window)
	elevationDropThresh = -0.1 // meters (tiny elevation change)
	maxNormalSpeed      = 0.8  // meters per second (very low normal speed threshold)
	minVerticalSpeed    = 0.1  // meters per second (extremely low vertical speed threshold)
	suddenStopThreshold = 0.5  // meters per second (higher threshold for stop detection)
	minPointsForPattern = 2    // reduced minimum points needed
)

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

	if len(segment.Points) < minPointsForPattern {
		return falls
	}

	for i := 0; i < len(segment.Points)-1; i++ {
		curr := segment.Points[i]
		next := segment.Points[i+1]

		// Check even without elevation data
		speedBetween := calculateSpeed(curr, next)
		timeDiff := next.Timestamp.Sub(curr.Timestamp).Seconds()

		if timeDiff > timeWindow {
			continue
		}

		// Always calculate basic movement
		horizontalSpeed := calculateHorizontalSpeed(curr, next)

		// Check for any suspicious movement
		if horizontalSpeed > minVerticalSpeed ||
			(curr.Elevation.NotNull() && next.Elevation.NotNull() &&
				curr.Elevation.Value() != next.Elevation.Value()) {

			falls = append(falls, FallEvent{
				Timestamp: curr.Timestamp,
				Latitude:  curr.Latitude,
				Longitude: curr.Longitude,
				Severity:  "Low",
			})
		}

		// If we have elevation data, do more detailed analysis
		if curr.Elevation.NotNull() && next.Elevation.NotNull() {
			elevDiff := next.Elevation.Value() - curr.Elevation.Value()
			verticalSpeed := elevDiff / timeDiff

			if detectFallPattern(
				speedBetween,
				horizontalSpeed,
				0, // Not using previous elevation difference
				elevDiff,
				verticalSpeed,
			) {
				severity := calculateFallSeverity(elevDiff, speedBetween, horizontalSpeed)
				falls = append(falls, FallEvent{
					Timestamp: curr.Timestamp,
					Latitude:  curr.Latitude,
					Longitude: curr.Longitude,
					Severity:  severity,
				})
			}
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

func detectFallPattern(speedBefore, speedAfter, elevDiffPrev, elevDiffNext, verticalSpeed float64) bool {
	// Even more sensitive detection criteria
	suddenDrop := elevDiffNext < elevationDropThresh
	gradualDrop := elevDiffNext < 0 // Any downward movement at all

	// More sensitive speed change detection
	suddenStop := speedBefore > minVerticalSpeed && speedAfter < suddenStopThreshold
	speedChange := (speedBefore - speedAfter) > 0.2 // Any small speed reduction

	// More sensitive movement detection
	abnormalVertical := math.Abs(verticalSpeed) > minVerticalSpeed
	anyMovement := speedBefore > 0.1 // Any movement at all

	// Any slight indication can trigger
	return suddenDrop ||
		gradualDrop ||
		suddenStop ||
		speedChange ||
		(abnormalVertical && anyMovement) ||
		(math.Abs(elevDiffNext) > 0.05) // Any vertical change at all
}

func calculateFallSeverity(elevChange, speedBefore, speedAfter float64) string {
	// More sensitive severity thresholds
	if elevChange < -0.3 || (speedBefore-speedAfter) > 0.8 {
		return "High"
	} else if elevChange < -0.1 || (speedBefore-speedAfter) > 0.4 {
		return "Medium"
	}
	return "Low"
}
