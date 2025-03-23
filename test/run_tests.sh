#!/bin/bash

# Simple test runner script
echo "Running Docker Compose tests..."

# Change to test directory
cd "$(dirname "$0")" || exit

# Run Go tests
go test -v ./...

echo "Tests completed"