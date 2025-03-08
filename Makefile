.PHONY: nothing test lint

nothing:
	@echo -n 'bagend '
	@git rev-parse HEAD

test:
	go test ./go/...

lint:
	golangci-lint run ./go/...
