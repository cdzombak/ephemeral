.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: bootstrap
bootstrap: ## Grab dependencies
	go get github.com/ChimeraCoder/anaconda
	go get github.com/aws/aws-lambda-go/lambda
	go get github.com/crewjam/errset
	go get github.com/pkg/errors

dist: bootstrap main.go ## Build the ephemeral binary & distribution archive
	golint -set_exit_status *.go
	gofmt -e -w *.go
	goimports -w *.go
	mkdir -p dist
	GOOS=linux go build -o dist/ephemeral
	zip -j dist/ephemeral.zip dist/ephemeral

.PHONY: deploy
deploy: dist ## Deploy built product to AWS under the name ephemeral-$EPHEMERAL_INSTANCE_NAME
	bash ./deploy.sh

.PHONY: clean
clean: ## Clean built products and any temporary files
	rm -rf dist
