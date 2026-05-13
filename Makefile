include .env
export

.PHONY: deploy destroy plan dev tf-output build-authorizer create-user authenticate-user oauth-user

deploy:
	cd terraform && terraform apply -auto-approve

destroy:
	cd terraform && terraform destroy -auto-approve

plan:
	cd terraform && terraform plan

tf-output:
	cd terraform && terraform output

dev:
	APPSYNC_API_KEY=$(APPSYNC_API_KEY) \
	HTTP_ENDPOINT=$(HTTP_ENDPOINT) \
	WS_ENDPOINT=$(WS_ENDPOINT) \
	CHANNEL=$(CHANNEL) \
	AWS_REGION=$(AWS_REGION) \
	ID_TOKEN=$(ID_TOKEN) \
	OIDC_TOKEN=$(OIDC_TOKEN) \
	go run ./internal/cmd/dev/

build-authorizer:
	mkdir -p dist/authorizer
	GOOS=linux GOARCH=arm64 go build -o dist/authorizer/bootstrap ./internal/cmd/authorizer/
	cd dist/authorizer && zip -j function.zip bootstrap

create-user:
ifndef USERNAME
	$(error USERNAME is not set. Usage: make create-user USERNAME=<user> PASSWORD=<pass>)
endif
ifndef PASSWORD
	$(error PASSWORD is not set. Usage: make create-user USERNAME=<user> PASSWORD=<pass>)
endif
	bash scripts/create-user.sh "$(USERNAME)" "$(PASSWORD)"
	$(MAKE) authenticate-user

authenticate-user:
ifndef USERNAME
	$(error USERNAME is not set. Usage: make authenticate-user USERNAME=<user> PASSWORD=<pass>)
endif
ifndef PASSWORD
	$(error PASSWORD is not set. Usage: make authenticate-user USERNAME=<user> PASSWORD=<pass>)
endif
	bash scripts/authenticate-user.sh "$(USERNAME)" "$(PASSWORD)"

oauth-user:
	bash scripts/oauth-user.sh

