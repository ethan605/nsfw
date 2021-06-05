.PHONY: crawler

PROJECT_NAME = nsfw

test:
	go test ./... -v -race -coverprofile=coverage.txt -covermode=atomic

crawler:
	docker-compose \
		--file deployments/docker-compose.yml \
		--project-name $(PROJECT_NAME) \
		up --build $@
