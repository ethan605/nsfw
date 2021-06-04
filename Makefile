test:
	go test ./... -v -race -coverprofile=coverage.txt -covermode=atomic

crawler:
	cd cmd/crawler && go run -race .
