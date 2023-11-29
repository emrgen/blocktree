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
