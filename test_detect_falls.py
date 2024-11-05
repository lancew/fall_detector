import unittest
from detect_falls import detect_falls

class TestFallDetection(unittest.TestCase):
    def test_specific_location_fall(self):
        """Test if a fall is detected near 50°58'22.7"N 1°15'21.8"W"""
        falls = detect_falls("Fishers_Pond__Owslesbury__Longwood__Uphams_-2.gpx")
        
        # Convert coordinates to decimal format
        target_lat = 50.97297
        target_lon = -1.25606
        
        # Check if any detected falls are within 0.0001 degrees (roughly 10m) of target
        falls_near_target = [
            fall for fall in falls
            if abs(fall['latitude'] - target_lat) < 0.0001 
            and abs(fall['longitude'] - target_lon) < 0.0001
        ]
        
        self.assertTrue(
            len(falls_near_target) > 0,
            "No falls detected near coordinates 50°58'22.7\"N 1°15'21.8\"W"
        )

if __name__ == '__main__':
    unittest.main()
