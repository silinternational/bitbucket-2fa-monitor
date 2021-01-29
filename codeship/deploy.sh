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
export AWS_REGION="${AWS_REGION}"
export AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID}"
export AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY}"

# Deploy
[[ -z "$1" ]] && { echo "Error: Environment not specified"; exit 1; }
serverless deploy -v --stage $1