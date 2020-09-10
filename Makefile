mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
mkfile_dir := $(dir $(mkfile_path))

.PHONY: all
all: lint test install

.PHONY: install
install:
    # -s omits the symbol table and debug information
    # -w omits the DWARF symbol table
	go install -ldflags "-s -w"

.PHONY: lint
lint:
	echo $(mkfile_dir)
	docker run --rm -v $(mkfile_dir):/app -w /app golangci/golangci-lint:v1.30.0 golangci-lint run -v -E goimports,unconvert,misspell,gocyclo,deadcode,errcheck,gosimple,govet,ineffassign,staticcheck,structcheck,typecheck,unused,varcheck,gocritic,gochecknoinits

.PHONY: test
test:
	go test -race ./...

.PHONY: test
release:
	env GOOS=linux GOARCH=amd64 go build -o releases/linux_amd64/noteo
	cd releases/linux_amd64 && tar cfz noteo_linux.tar.gz noteo
	env GOOS=darwin GOARCH=amd64 go build -o releases/darwin_amd64/noteo
	cd releases/darwin_amd64 && tar cfz noteo_macos.tar.gz noteo
	env GOOS=windows GOARCH=amd64 go build -o releases/windows_amd64/noteo
	cd releases/windows_amd64 && zip noteo_windows.zip noteo