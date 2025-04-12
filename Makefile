CLIENT_VERSION = $(shell cat ./version | grep  "client-version=*" | awk -F"=" '{ print $$2 }')

.PHONY: init
init: proto deps docs
	@echo "blocktree setup complete"
	@echo "to start the server run: make start"

.PHONY: start
start:
	@echo "Starting blocktree server"
	@go run cmd/cli/main.go serve

.PHONY: build
build:
	@echo "Building blocktree server"
	@go build -o ./bin/bt ./cmd/cli/main.go

.PHONY: run
air:
	@echo "Starting blocktree server with air"
	@air

.PHONY: test
test:
	@echo "Running tests..."
	go test

.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod vendor

.PHONY: protoc
protoc:
	@echo "Generating proto files..."
	buf generate


clean:
	@echo "Cleaning..."
	rm -rf ./bin
	rm -rf ./vendor
	rm -rf ./apis

generate-ts-client:
	@echo "Generating typescript client..."
	@openapi-generator-cli generate -i ./apis/v1/blocktree.swagger.json \
		-g typescript-axios -o ./client/blocktree-ts-client \
		--additional-properties=npmName=@emrgen/blocktree-client,npmVersion=$(CLIENT_VERSION),useSingleRequestParameter=true \
        --type-mappings=string=String

.PHONY: client
client: proto generate-ts-client

.PHONY: godoc
doc:
	google-chrome http://localhost:6060/pkg/github.com/emrgen/blocktree
	godoc -http=:6060

.PHONY: docs
docs:
	@echo "Generating API documentation..."
	@mkdir -p ./docs
	@npx @redocly/cli build-docs ./apis/v1/blocktree.swagger.json --output ./docs/v1/index.html

.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run