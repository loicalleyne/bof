BINARY := bof-mcp
MODULE := github.com/loicalleyne/bof/bof-mcp
INSTALL_DIR := $(HOME)/.local/bin

.PHONY: build test test-race vet install tidy clean

build:
	go build -o $(BINARY) .

test:
	go test -count=1 -timeout 60s ./...

test-race:
	go test -count=1 -race -timeout 120s -run '^Test' ./...

vet:
	go vet ./...

install: build
	install -m 0755 $(BINARY) $(INSTALL_DIR)/$(BINARY)

tidy:
	go mod tidy

clean:
	rm -f $(BINARY)
