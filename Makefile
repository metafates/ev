test:
	go test -v -race .

lint:
	golangci-lint run

fmt:
	golangci-lint fmt
