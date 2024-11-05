import xml.etree.ElementTree as ET
from datetime import datetime
from dataclasses import dataclass
from typing import List, Optional
import math

@dataclass
class TrackPoint:
    latitude: float
    longitude: float
    elevation: float
    timestamp: datetime

@dataclass
class FallEvent:
    timestamp: datetime
    latitude: float
    longitude: float
    elevation_drop: float
    speed: float
    duration_into_ride: float  # New field: seconds since ride start
    stationary: bool = False

class GPXAnalyzer:
    def __init__(self, 
                 min_elevation_drop: float = -2.0,    # Changed from -1.5 to -2.0 meters (less sensitive)
                 time_window: float = 5.0,            # Changed from 1.5 to 2.0 seconds (more forgiving)
                 min_speed: float = 1.5,              # Changed from 1.5 to 1.0 m/s (more sensitive)
                 stationary_radius: float = 3.0,      # Changed from 4.0 to 3.0 meters (more sensitive)
                 stationary_time: float = 60.0,       # Changed from 90.0 to 60.0 seconds (more sensitive)
                 max_speed: float = 30.0,             # Kept the same
                 start_exclusion_window: float = 300.0,
                 end_exclusion_window: float = 300.0):
        self.min_elevation_drop = min_elevation_drop
        self.time_window = time_window
        self.min_speed = min_speed
        self.stationary_radius = stationary_radius
        self.stationary_time = stationary_time
        self.max_speed = max_speed
        self.start_exclusion_window = start_exclusion_window
        self.end_exclusion_window = end_exclusion_window

    def parse_gpx(self, filename: str) -> List[TrackPoint]:
        tree = ET.parse(filename)
        root = tree.getroot()
        
        # Handle namespace in GPX files
        ns = {'gpx': 'http://www.topografix.com/GPX/1/1'}
        
        points = []
        for trkpt in root.findall('.//gpx:trkpt', ns):
            lat = float(trkpt.get('lat'))
            lon = float(trkpt.get('lon'))
            ele = float(trkpt.find('gpx:ele', ns).text)
            time_str = trkpt.find('gpx:time', ns).text
            timestamp = datetime.fromisoformat(time_str.replace('Z', '+00:00'))
            
            points.append(TrackPoint(lat, lon, ele, timestamp))
            
        return points

    def calculate_speed(self, p1: TrackPoint, p2: TrackPoint) -> float:
        """Calculate speed between two points in meters per second"""
        R = 6371000  # Earth radius in meters
        
        lat1, lon1 = math.radians(p1.latitude), math.radians(p1.longitude)
        lat2, lon2 = math.radians(p2.latitude), math.radians(p2.longitude)
        
        dlat = lat2 - lat1
        dlon = lon2 - lon1
        
        a = math.sin(dlat/2)**2 + math.cos(lat1) * math.cos(lat2) * math.sin(dlon/2)**2
        c = 2 * math.atan2(math.sqrt(a), math.sqrt(1-a))
        distance = R * c
        
        time_diff = (p2.timestamp - p1.timestamp).total_seconds()
        if time_diff == 0:
            return 0
        return distance / time_diff

    def filter_outliers(self, points: List[TrackPoint]) -> List[TrackPoint]:
        """Remove points that would result in unrealistic speeds."""
        if len(points) <= 1:
            return points
            
        filtered_points = [points[0]]
        
        for i in range(1, len(points)):
            speed = self.calculate_speed(filtered_points[-1], points[i])
            if speed <= self.max_speed:
                filtered_points.append(points[i])
                
        return filtered_points

    def detect_falls(self, filename: str) -> List[FallEvent]:
        points = self.parse_gpx(filename)
        points = self.filter_outliers(points)  # Filter out outlier points
        falls = []
        
        if not points:
            return falls
            
        ride_start_time = points[0].timestamp  # Store the start time of the ride
        ride_end_time = points[-1].timestamp
        total_duration = (ride_end_time - ride_start_time).total_seconds()
        
        # Check for stationary periods
        i = 0
        while i < len(points):
            curr = points[i]
            stationary_start = curr.timestamp
            last_moving_idx = i
            
            # Look ahead to check if person stays within radius
            for j in range(i + 1, len(points)):
                if self.calculate_speed(curr, points[j]) * \
                   (points[j].timestamp - curr.timestamp).total_seconds() > self.stationary_radius:
                    last_moving_idx = j
                    break
            
            # If person hasn't moved far and time exceeds threshold, record as fall
            time_diff = (points[last_moving_idx].timestamp - stationary_start).total_seconds()
            if time_diff >= self.stationary_time:
                duration_into_ride = (stationary_start - ride_start_time).total_seconds()
                # Only add fall if it's not in exclusion windows
                if (duration_into_ride > self.start_exclusion_window and 
                    duration_into_ride < (total_duration - self.end_exclusion_window)):
                    falls.append(FallEvent(
                        timestamp=stationary_start,
                        latitude=curr.latitude,
                        longitude=curr.longitude,
                        elevation_drop=0.0,
                        speed=0.0,
                        duration_into_ride=duration_into_ride,
                        stationary=True
                    ))
                i = last_moving_idx
            
            # Check for elevation-based falls
            if i < len(points) - 1:
                next_point = points[i + 1]
                time_diff = (next_point.timestamp - curr.timestamp).total_seconds()
                
                if time_diff <= self.time_window:
                    elevation_change = next_point.elevation - curr.elevation
                    speed = self.calculate_speed(curr, next_point)
                    
                    if (elevation_change < self.min_elevation_drop and 
                        speed > self.min_speed):
                        duration_into_ride = (curr.timestamp - ride_start_time).total_seconds()
                        # Only add fall if it's not in exclusion windows
                        if (duration_into_ride > self.start_exclusion_window and 
                            duration_into_ride < (total_duration - self.end_exclusion_window)):
                            falls.append(FallEvent(
                                timestamp=curr.timestamp,
                                latitude=curr.latitude,
                                longitude=curr.longitude,
                                elevation_drop=elevation_change,
                                speed=speed,
                                duration_into_ride=duration_into_ride,
                                stationary=False
                            ))
            i += 1
        
        return falls

def main():
    analyzer = GPXAnalyzer()
    
    # Get filename from command line argument or use default
    import sys
    filename = sys.argv[1] if len(sys.argv) > 1 else "track.gpx"
    
    try:
        falls = analyzer.detect_falls(filename)
        
        if not falls:
            print("No potential falls detected")
            return
            
        print(f"Detected {len(falls)} potential falls:")
        for i, fall in enumerate(falls, 1):
            print(f"\nFall {i}:")
            print(f"Time: {fall.timestamp}")
            print(f"Duration into ride: {fall.duration_into_ride/60:.1f} minutes")  # Convert to minutes
            print(f"Location: ({fall.latitude}, {fall.longitude})")
            if fall.stationary:
                print("Type: Stationary incident")
            else:
                print(f"Elevation drop: {fall.elevation_drop:.2f}m")
                print(f"Speed: {fall.speed:.2f}m/s")

    except Exception as e:
        print(f"Error analyzing GPX file: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()