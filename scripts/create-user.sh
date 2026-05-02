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

echo "Creating user '$USERNAME'..."

CREATE_OUTPUT=$(aws cognito-idp admin-create-user \
  --region "$AWS_REGION" \
  --user-pool-id "$USER_POOL" \
  --username "$USERNAME" \
  --message-action SUPPRESS \
  2>&1) && USER_CREATED=true || USER_CREATED=false

if ! $USER_CREATED; then
  if echo "$CREATE_OUTPUT" | grep -q "UsernameExistsException"; then
    echo "User '$USERNAME' already exists — skipping creation."
  else
    echo "Error creating user:" >&2
    echo "$CREATE_OUTPUT" >&2
    exit 1
  fi
else
  echo "User '$USERNAME' created."
fi

echo "Setting permanent password..."
aws cognito-idp admin-set-user-password \
  --region "$AWS_REGION" \
  --user-pool-id "$USER_POOL" \
  --username "$USERNAME" \
  --password "$PASSWORD" \
  --permanent

echo "user created."
