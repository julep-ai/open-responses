#!/usr/bin/env python
import os
from setuptools import setup, find_packages

# Reading the package version from pyproject.toml
from importlib.metadata import version, PackageNotFoundError

try:
    __version__ = version('open-responses')
except PackageNotFoundError:
    __version__ = '0.1.0'  # Default version if not installed

setup(
    name="open-responses",
    version=__version__,
    description="CLI for setting up a self-hosted alternative to OpenAI's Responses API",
    author="Julep AI",
    author_email="info@example.com",
    url="https://github.com/julep-ai/open-responses",
    packages=find_packages(),
    include_package_data=True,
    entry_points={
        'console_scripts': [
            'open-responses=open_responses:main',
        ],
    },
    install_requires=[],
    classifiers=[
        "Programming Language :: Python :: 3",
        "License :: OSI Approved :: Apache Software License",
        "Operating System :: OS Independent",
    ],
    python_requires=">=3.7",
)