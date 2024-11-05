#!/usr/bin/env python3

import sys
import gpxpy
import numpy as np
from datetime import datetime

def calculate_speed(point1, point2):
    """Calculate speed between two GPX points in m/s"""
    time_diff = (point2.time - point1.time).total_seconds()
    if time_diff == 0:
        return 0
    
    distance = point1.distance_3d(point2)
    return distance / time_diff

def detect_falls(gpx_file):
    """Detect potential falls in GPX track"""
    with open(gpx_file, 'r') as f:
        gpx = gpxpy.parse(f)
    
    falls = []
    
    for track in gpx.tracks:
        for segment in track.segments:
            points = segment.points
            if len(points) < 3:
                continue
                
            # Calculate speeds and elevation changes
            for i in range(1, len(points)-1):
                prev_point = points[i-1]
                curr_point = points[i]
                next_point = points[i+1]
                
                # Calculate speeds
                speed_before = calculate_speed(prev_point, curr_point)
                speed_after = calculate_speed(curr_point, next_point)
                
                # Calculate elevation changes
                elev_change = curr_point.elevation - prev_point.elevation
                
                # Detect sudden speed drops and elevation changes
                speed_drop = speed_before - speed_after
                
                # Thresholds for fall detection
                if (speed_drop > 10.0 and  # Speed drop > 10 m/s
                    abs(elev_change) > 2.5 and  # Elevation change > 2.5m
                    speed_before > 5.0):  # Initial speed > 5 m/s
                    
                    falls.append({
                        'time': curr_point.time,
                        'latitude': curr_point.latitude,
                        'longitude': curr_point.longitude,
                        'speed_drop': speed_drop,
                        'elevation_change': elev_change
                    })
    
    return falls

def main():
    if len(sys.argv) != 2:
        print("Usage: python3 detect_falls.py <gpx_file>")
        sys.exit(1)
        
    gpx_file = sys.argv[1]
    falls = detect_falls(gpx_file)
    
    if not falls:
        print("No falls detected")
    else:
        print(f"Detected {len(falls)} potential falls:")
        for fall in falls:
            print(f"\nTime: {fall['time']}")
            print(f"Location: {fall['latitude']}, {fall['longitude']}")
            print(f"Speed drop: {fall['speed_drop']:.2f} m/s")
            print(f"Elevation change: {fall['elevation_change']:.2f} m")

if __name__ == "__main__":
    main()
