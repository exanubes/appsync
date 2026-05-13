#!/usr/bin/env bash
set -euo pipefail

AWS_REGION="eu-central-1"
LAMBDA_AUTHORIZER_TOKEN="local-e2e-token"
E2E_USERNAME="local-e2e-user"
E2E_PASSWORD="local-e2e-pass"

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TF_DIR="$REPO_ROOT/terraform"
ENV_FILE="$REPO_ROOT/.env.e2e"

echo "==> Building Lambda authorizer..."
mkdir -p "$REPO_ROOT/dist/authorizer"
GOOS=linux GOARCH=arm64 go build -o "$REPO_ROOT/dist/authorizer/bootstrap" "$REPO_ROOT/internal/cmd/authorizer/"
cd "$REPO_ROOT/dist/authorizer" && zip -j function.zip bootstrap
cd "$REPO_ROOT"

echo "==> Applying Terraform..."
terraform -chdir="$TF_DIR" apply \
  -auto-approve \
  -input=false \
  -var="aws_region=$AWS_REGION" \
  -var="lambda_authorizer_token=$LAMBDA_AUTHORIZER_TOKEN"

echo "==> Reading Terraform outputs..."
HTTP_ENDPOINT="$(terraform -chdir="$TF_DIR" output -raw appsync_http_endpoint)"
WS_ENDPOINT="$(terraform -chdir="$TF_DIR" output -raw appsync_ws_endpoint)"
API_KEY="$(terraform -chdir="$TF_DIR" output -raw appsync_api_key)"
USER_POOL_ID="$(terraform -chdir="$TF_DIR" output -raw user_pool_id)"
COGNITO_CLIENT_ID="$(terraform -chdir="$TF_DIR" output -raw cognito_user_pool_client_id)"
COGNITO_DOMAIN="$(terraform -chdir="$TF_DIR" output -raw cognito_domain)"
OIDC_CLIENT_ID="$(terraform -chdir="$TF_DIR" output -raw oidc_user_pool_client_id)"
OIDC_CLIENT_SECRET="$(terraform -chdir="$TF_DIR" output -raw oidc_user_pool_client_secret)"
OIDC_SCOPE="$(terraform -chdir="$TF_DIR" output -raw oidc_scope)"
NS_API_KEY="$(terraform -chdir="$TF_DIR" output -raw namespace_api_key)"
NS_IAM="$(terraform -chdir="$TF_DIR" output -raw namespace_iam)"
NS_LAMBDA="$(terraform -chdir="$TF_DIR" output -raw namespace_lambda)"
NS_COGNITO="$(terraform -chdir="$TF_DIR" output -raw namespace_cognito)"
NS_OIDC="$(terraform -chdir="$TF_DIR" output -raw namespace_oidc)"

echo "==> Creating Cognito user '$E2E_USERNAME'..."
aws cognito-idp admin-create-user \
  --region "$AWS_REGION" \
  --user-pool-id "$USER_POOL_ID" \
  --username "$E2E_USERNAME" \
  --message-action SUPPRESS 2>&1 | grep -v UsernameExistsException || true

aws cognito-idp admin-set-user-password \
  --region "$AWS_REGION" \
  --user-pool-id "$USER_POOL_ID" \
  --username "$E2E_USERNAME" \
  --password "$E2E_PASSWORD" \
  --permanent

echo "==> Getting Cognito ID token..."
AUTH_RESULT="$(aws cognito-idp initiate-auth \
  --region "$AWS_REGION" \
  --auth-flow USER_PASSWORD_AUTH \
  --client-id "$COGNITO_CLIENT_ID" \
  --auth-parameters "USERNAME=$E2E_USERNAME,PASSWORD=$E2E_PASSWORD")"

COGNITO_ID_TOKEN="$(echo "$AUTH_RESULT" | jq -r '.AuthenticationResult.IdToken')"
if [[ -z "$COGNITO_ID_TOKEN" || "$COGNITO_ID_TOKEN" == "null" ]]; then
  echo "Failed to obtain Cognito ID token" >&2
  echo "$AUTH_RESULT" >&2
  exit 1
fi

echo "==> Getting OIDC token..."
TOKEN_RESPONSE="$(curl -sf -X POST "${COGNITO_DOMAIN}/oauth2/token" \
  -u "${OIDC_CLIENT_ID}:${OIDC_CLIENT_SECRET}" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials&scope=${OIDC_SCOPE}")"

OIDC_TOKEN="$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')"
if [[ -z "$OIDC_TOKEN" || "$OIDC_TOKEN" == "null" ]]; then
  echo "Failed to obtain OIDC access token" >&2
  echo "$TOKEN_RESPONSE" >&2
  exit 1
fi

echo "==> Writing $ENV_FILE..."
cat > "$ENV_FILE" <<EOF
AWS_REGION="$AWS_REGION"
APPSYNC_E2E_HTTP_ENDPOINT="$HTTP_ENDPOINT"
APPSYNC_E2E_WS_ENDPOINT="$WS_ENDPOINT"
APPSYNC_E2E_API_KEY="$API_KEY"
APPSYNC_E2E_LAMBDA_TOKEN="$LAMBDA_AUTHORIZER_TOKEN"
APPSYNC_E2E_COGNITO_ID_TOKEN="$COGNITO_ID_TOKEN"
APPSYNC_E2E_OIDC_TOKEN="$OIDC_TOKEN"
APPSYNC_E2E_NS_API_KEY="$NS_API_KEY"
APPSYNC_E2E_NS_IAM="$NS_IAM"
APPSYNC_E2E_NS_LAMBDA="$NS_LAMBDA"
APPSYNC_E2E_NS_COGNITO="$NS_COGNITO"
APPSYNC_E2E_NS_OIDC="$NS_OIDC"
EOF

echo "==> Setup complete. Run 'make e2e-test' to run the tests."
