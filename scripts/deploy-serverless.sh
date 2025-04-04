#!/bin/bash

# Simple Serverless Deployment Script using AWS SAM CLI

# Exit immediately if a command exits with a non-zero status.
set -e

# --- Configuration ---
APP_NAME="microblog-platform"
TEMPLATE_FILE="infrastructure/aws/template.yaml"
# Default AWS Region (can be overridden by environment variable AWS_DEFAULT_REGION)
AWS_REGION=${AWS_DEFAULT_REGION:-"us-east-1"}

# --- Helper Functions ---
usage() {
  echo "Usage: $0 <environment> <redis_endpoint_address> [redis_endpoint_port]"
  echo "  environment: Deployment environment (e.g., dev, staging, prod). Used for stack name."
  echo "  redis_endpoint_address: The endpoint address of the ElastiCache Redis instance."
  echo "  redis_endpoint_port: The port of the ElastiCache Redis instance (default: 6379)."
  exit 1
}

# --- Script Logic ---

# Validate arguments
if [ "$#" -lt 2 ] || [ "$#" -gt 3 ]; then
  usage
fi

ENVIRONMENT=$1
REDIS_ENDPOINT_ADDRESS=$2
REDIS_ENDPOINT_PORT=${3:-6379} # Default to 6379 if not provided

STACK_NAME="${APP_NAME}-${ENVIRONMENT}"

echo "--- Deployment Configuration ---"
echo "Environment:            ${ENVIRONMENT}"
echo "Stack Name:             ${STACK_NAME}"
echo "AWS Region:             ${AWS_REGION}"
echo "SAM Template:           ${TEMPLATE_FILE}"
echo "Redis Endpoint Address: ${REDIS_ENDPOINT_ADDRESS}"
echo "Redis Endpoint Port:    ${REDIS_ENDPOINT_PORT}"
echo "------------------------------"

# 1. Build the Go Binary for Lambda (Linux AMD64)
# Assumes script is run from the project root directory
echo "\n---> Building Go binary..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main ./cmd/main.go
echo "Build complete."

# 2. Build the SAM deployment package
echo "\n---> Building SAM package..."
# Run sam build from the root directory so CodeUri: ../../ works relative to template.yaml
sam build --template ${TEMPLATE_FILE} --use-container # Use --use-container for consistency
echo "SAM build complete."

# 3. Deploy the application using SAM
echo "\n---> Deploying SAM application..."
sam deploy \
  --stack-name ${STACK_NAME} \
  --region ${AWS_REGION} \
  --capabilities CAPABILITY_IAM CAPABILITY_AUTO_EXPAND \
  --parameter-overrides "RedisEndpointAddress=${REDIS_ENDPOINT_ADDRESS} RedisEndpointPort=${REDIS_ENDPOINT_PORT}" \
  --resolve-s3 # Use SAM managed S3 bucket for artifacts
  # --guided # Uncomment for initial deployment or if you want guided prompts

echo "\nDeployment script finished successfully!"

# Optional: Clean up the built binary
# echo "\n---> Cleaning up build artifact..."
# rm main

exit 0 