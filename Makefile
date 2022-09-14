MD_FILES := $(shell find . -type f -regex ".*md"  -not -regex '^./vendor/.*'  -not -regex '^./.vale/.*' -not -regex "^./.git/.*" -print)
BINARY_NAME := cumin

LDFLAGS := -s -w
FLAGS += -ldflags "$(LDFLAGS)" -buildvcs=true

all: lint build

clean:
	@rm -rf bin/gosmee

build: clean
	@echo "building $(BINARY_NAME) to bin/$(BINARY_NAME)"
	@mkdir -p bin/
	@go build  -v $(FLAGS)  -o bin/$(BINARY_NAME) main.go

lint: lint-go lint-md

lint-go:
	@echo "linting."
	@golangci-lint run --disable gosimple --disable staticcheck --disable structcheck --disable unused

.PHONY: lint-md
lint-md: ${MD_FILES} ## runs markdownlint and vale on all markdown files
	@echo "Linting markdown files..."
	@markdownlint $(MD_FILES)

fmt:
	@go fmt `go list ./... | grep -v /vendor/`

fumpt:
	@gofumpt -w *.go

.PHONY: vendor
vendor:
	@go mod tidy
	@go mod vendor
