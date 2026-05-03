#!/usr/bin/env bash
set -euo pipefail

: "${OIDC_CLIENT_ID:?OIDC_CLIENT_ID must be set in .env}"
: "${USER_POOL:?USER_POOL must be set in .env}"
: "${AWS_REGION:?AWS_REGION must be set in .env}"

OIDC_CLIENT_ID="${OIDC_CLIENT_ID//\"/}"
USER_POOL="${USER_POOL//\"/}"
AWS_REGION="${AWS_REGION//\"/}"

ENV_FILE="$(dirname "$0")/../.env"

OIDC_CLIENT_SECRET=$(aws cognito-idp describe-user-pool-client \
  --region "$AWS_REGION" \
  --user-pool-id "$USER_POOL" \
  --client-id "$OIDC_CLIENT_ID" \
  --query 'UserPoolClient.ClientSecret' \
  --output text)

DISCOVERY_URL="https://cognito-idp.${AWS_REGION}.amazonaws.com/${USER_POOL}/.well-known/openid-configuration"
OIDC_TOKEN_ENDPOINT=$(curl -sf "$DISCOVERY_URL" | jq -r '.token_endpoint')

if [[ "$OIDC_TOKEN_ENDPOINT" == "null" || -z "$OIDC_TOKEN_ENDPOINT" ]]; then
  echo "Could not resolve token_endpoint from $DISCOVERY_URL" >&2
  exit 1
fi

echo "Fetching OIDC token via client_credentials from $OIDC_TOKEN_ENDPOINT..."
RESPONSE=$(curl -s -X POST "$OIDC_TOKEN_ENDPOINT" \
  -u "${OIDC_CLIENT_ID}:${OIDC_CLIENT_SECRET}" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials&scope=exanubes%2Fcustom")

OIDC_TOKEN=$(echo "$RESPONSE" | jq -r '.access_token')

if [[ "$OIDC_TOKEN" == "null" || -z "$OIDC_TOKEN" ]]; then
  echo "Authentication failed — no access_token in response." >&2
  echo "$RESPONSE" >&2
  exit 1
fi

set_env_var() {
  local key="$1" value="$2" file="$3"
  if grep -qE "^${key}=" "$file"; then
    sed -i '' "s|^${key}=.*|${key}=\"${value}\"|" "$file"
  else
    echo "${key}=\"${value}\"" >> "$file"
  fi
}

set_env_var "OIDC_TOKEN" "$OIDC_TOKEN" "$ENV_FILE"

echo ""
echo "Token saved to .env:"
echo "  OIDC_TOKEN"
