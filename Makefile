

tidy:
	go mod tidy

lint:
	golangci-lint run --verbose
