#!/bin/bash

# Create virtual environment if it doesn't exist
if [ ! -d "venv" ]; then
    python3 -m venv venv
fi

# Activate virtual environment
source venv/bin/activate

# Install requirements
pip install -r requirements.txt

# Run the script with the sample GPX file
python detect_falls.py Fishers_Pond__Owslesbury__Longwood__Uphams_-2.gpx

# Deactivate virtual environment
deactivate
