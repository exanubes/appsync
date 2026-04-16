include .env

.PHONY: deploy destroy plan dev tf-output

deploy:
	cd terraform && terraform apply -auto-approve

destroy:
	cd terraform && terraform destroy -auto-approve

plan:
	cd terraform && terraform plan

tf-output:
	cd terraform && terraform output

dev:
	HTTP_ENDPOINT=$(HTTP_ENDPOINT) WS_ENDPOINT=$(WS_ENDPOINT) CHANNEL=$(CHANNEL) AWS_REGION=$(AWS_REGION) go run ./cmd/dev/
