include .env

.PHONY: deploy destroy plan dev tf-output build-authorizer

deploy:
	cd terraform && terraform apply -auto-approve

destroy:
	cd terraform && terraform destroy -auto-approve

plan:
	cd terraform && terraform plan

tf-output:
	cd terraform && terraform output

dev:
	APPSYNC_API_KEY=$(APPSYNC_API_KEY) HTTP_ENDPOINT=$(HTTP_ENDPOINT) WS_ENDPOINT=$(WS_ENDPOINT) CHANNEL=$(CHANNEL) AWS_REGION=$(AWS_REGION) go run ./cmd/dev/

build-authorizer:
	mkdir -p dist/authorizer
	GOOS=linux GOARCH=arm64 go build -o dist/authorizer/bootstrap ./cmd/authorizer/
	cd dist/authorizer && zip -j function.zip bootstrap
