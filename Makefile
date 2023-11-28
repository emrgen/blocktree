
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod vendor

clean:
	@echo "Cleaning..."
	rm -rf ./bin
	rm -rf ./vendor
