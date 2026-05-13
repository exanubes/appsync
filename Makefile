.PHONY: build-authorizer e2e e2e-setup e2e-test e2e-teardown

build-authorizer:
	mkdir -p dist/authorizer
	GOOS=linux GOARCH=arm64 go build -o dist/authorizer/bootstrap ./internal/cmd/authorizer/
	cd dist/authorizer && zip -j function.zip bootstrap

e2e-setup:
	bash scripts/e2e-setup.sh

e2e-test:
	set -a && source .env.e2e && set +a && \
	go test -tags=e2e ./e2e -count=1 -timeout=3m -v

e2e-teardown:
	bash scripts/e2e-teardown.sh

e2e: e2e-setup
	$(MAKE) e2e-test; $(MAKE) e2e-teardown
