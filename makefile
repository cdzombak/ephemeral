.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: bootstrap
bootstrap: ## Grab dependencies
	go get github.com/ChimeraCoder/anaconda
	go get github.com/aws/aws-lambda-go/lambda

dist: bootstrap main.go ## Build the ephemeral binary & distribution archive
	mkdir -p dist
	GOOS=linux go build -o dist/ephemeral
	zip -j dist/ephemeral.zip dist/ephemeral

.PHONY: deploy
deploy: dist ## Deploy TODO CHECK DEPS AND TODO DOC ENVAR
	bash ./deploy.sh

.PHONY: clean
clean: ## Clean built products and any temporary files
	rm -rf dist
