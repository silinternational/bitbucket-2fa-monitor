#!/usr/bin/env bash

# Exit script with error if any step fails.
set -e

# Build binaries
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
$DIR/build.sh

# Export env vars
export API_BASE_URL="${API_BASE_URL}"
export API_USERNAME="${API_USERNAME}"
export API_APP_PASSWORD="${API_APP_PASSWORD}"
export API_WORKSPACE="${API_WORKSPACE}"
export SES_RETURN_TO_ADDRESS="${SES_RETURN_TO_ADDRESS}"
export SES_RECIPIENT_EMAILS="${SES_RECIPIENT_EMAILS}"

# Print the Serverless version in the logs
serverless --version

# Deploy
[[ -z "$1" ]] && { echo "Error: Environment not specified"; exit 1; }
echo "Deploying stage $1..."
serverless deploy --verbose --stage "$1"
