CLIENT_VERSION = $(shell cat ./version | grep  "client-version=*" | awk -F"=" '{ print $$2 }')

start:
	@echo "Starting blocktree server"
	@go run cmd/cli/main.go serve

build:
	@echo "Building blocktree server"
	@go build -o ./bin/bt ./cmd/cli/main.go

init: protoc deps
	@echo "blocktree setup complete"
	@echo "to start the server run: make start"

test:
	@echo "Running tests..."
	go test

deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod vendor

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
	@openapi-generator-cli generate -i ./apis/v1/apis.swagger.json \
		-g typescript-axios -o ./client/blocktree-ts-client \
		--additional-properties=npmName=@emrgen/blocktree-client,npmVersion=$(CLIENT_VERSION),useSingleRequestParameter=true \
        --type-mappings=string=String

client: protoc generate-ts-client

doc:
	google-chrome http://localhost:6060/pkg/github.com/emrgen/blocktree
	godoc -http=:6060
