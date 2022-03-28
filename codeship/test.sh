#!/usr/bin/env bash

# Exit script with error if any step fails.
set -e

# Echo commands to console
set -x

# Run Go tests
go test -v ./...

# Print the Serverless version in the logs
serverless --version

# Validate Serverless config
serverless info
