#!/bin/bash

# Ensure python3-distutils is installed (needed for numpy)
if ! dpkg -l | grep -q python3-distutils; then
    echo "Installing python3-distutils..."
    sudo pacman -S python-setuptools
fi

# Create and activate virtual environment
python3 -m venv .venv
source .venv/bin/activate

# Upgrade pip to latest version
pip install --upgrade pip

# Install requirements
pip install -r requirements.txt
