from setuptools import setup, find_packages

setup(
    name="fall-detector",
    version="0.1",
    packages=find_packages(),
    install_requires=[
        'gpxpy==1.5.0',
        'numpy==1.26.2',
    ],
    setup_requires=['pytest-runner'],
    tests_require=['pytest==7.4.3', 'gpxpy==1.5.0'],
)
