ifdef update
	u=-u
endif

export GO111MODULE=on

.PHONY: deps
deps:
	go get ${u} -d ./...
	go mod tidy

.PHONY: test
test:
	go test -race ./...