#!/usr/bin/env bash
set -euo pipefail

AWS_REGION="eu-central-1"
LAMBDA_AUTHORIZER_TOKEN="local-e2e-token"

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TF_DIR="$REPO_ROOT/terraform/e2e"
ENV_FILE="$REPO_ROOT/.env.e2e"

echo "==> Destroying Terraform resources..."
terraform -chdir="$TF_DIR" destroy \
  -auto-approve \
  -input=false \
  -var="aws_region=$AWS_REGION" \
  -var="lambda_authorizer_token=$LAMBDA_AUTHORIZER_TOKEN"

if [[ -f "$ENV_FILE" ]]; then
  trash "$ENV_FILE"
fi

echo "==> Teardown complete."
