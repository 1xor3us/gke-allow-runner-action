#!/usr/bin/env bash
set -e

# ──────────────────────────────────────────────
# CONFIGURATION - ADAPT TO YOUR ENVIRONMENT
# ──────────────────────────────────────────────
PROJECT_ID="reliable-plasma-476112-q3"
REGION="europe-west1"
CLUSTER_NAME="cluster-test"
SA_KEY_PATH="$(pwd)/ci-runner-access.json"
IMAGE_NAME="ghcr.io/1xor3us/gke-allow-runner-action:v1.0.0"
ACTION="allow"

echo
echo "GKE Allow GitHub Runner - Mode: $ACTION"
echo

docker run --rm \
  -v "$SA_KEY_PATH:/tmp/sa.json" \
  -e GOOGLE_APPLICATION_CREDENTIALS=/tmp/sa.json \
  -e INPUT_CLUSTER_NAME="$CLUSTER_NAME" \
  -e INPUT_REGION="$REGION" \
  -e INPUT_PROJECT_ID="$PROJECT_ID" \
  -e INPUT_ACTION="$ACTION" \
  "$IMAGE_NAME"

echo
echo "Script finished."
