#!/usr/bin/env bash
set -euo pipefail

USERNAME="${1:?USERNAME is required as first argument}"
PASSWORD="${2:?PASSWORD is required as second argument}"

: "${USER_POOL:?USER_POOL must be set in .env}"
: "${USER_POOL_CLIENT:?USER_POOL_CLIENT must be set in .env}"
: "${AWS_REGION:?AWS_REGION must be set in .env}"

# Make exports .env values with their surrounding quotes intact; strip them.
USER_POOL="${USER_POOL//\"/}"
USER_POOL_CLIENT="${USER_POOL_CLIENT//\"/}"
AWS_REGION="${AWS_REGION//\"/}"

ENV_FILE="$(dirname "$0")/../.env"
echo "Authenticating '$USERNAME'..."
AUTH_RESULT=$(aws cognito-idp initiate-auth \
  --region "$AWS_REGION" \
  --auth-flow USER_PASSWORD_AUTH \
  --client-id "$USER_POOL_CLIENT" \
  --auth-parameters "USERNAME=$USERNAME,PASSWORD=$PASSWORD")

ACCESS_TOKEN=$(echo "$AUTH_RESULT"  | jq -r '.AuthenticationResult.AccessToken')
ID_TOKEN=$(echo "$AUTH_RESULT"      | jq -r '.AuthenticationResult.IdToken')
REFRESH_TOKEN=$(echo "$AUTH_RESULT" | jq -r '.AuthenticationResult.RefreshToken')

if [[ "$ACCESS_TOKEN" == "null" || -z "$ACCESS_TOKEN" ]]; then
  echo "Authentication failed — no AccessToken in response." >&2
  echo "$AUTH_RESULT" >&2
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

# set_env_var "ACCESS_TOKEN"  "$ACCESS_TOKEN"  "$ENV_FILE"
set_env_var "ID_TOKEN"      "$ID_TOKEN"      "$ENV_FILE"
# set_env_var "REFRESH_TOKEN" "$REFRESH_TOKEN" "$ENV_FILE"

echo ""
echo "Tokens saved to .env:"
# echo "  ACCESS_TOKEN"
echo "  ID_TOKEN"
# echo "  REFRESH_TOKEN"
