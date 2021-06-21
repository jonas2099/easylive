export CGO_CXXFLAGS_ALLOW:=.*
export CGO_LDFLAGS_ALLOW:=.*
export CGO_CFLAGS_ALLOW:=.*

all: build
check:
	@echo "\033[32m <============== make check =============> \033[0m"
	gofmt -s -w .
	golangci-lint run