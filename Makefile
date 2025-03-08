.PHONY: nothing test lint

nothing:
	@echo -n 'bagend '
	@git rev-parse HEAD

test:
	go test ./go/...

lint:
	golangci-lint run ./go/...\
	  --max-issues-per-linter 0\
	  --max-same-issues 0\
	  --allow-parallel-runners\
	  --sort-results
