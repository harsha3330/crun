BINARY := crun
CMD_DIR := ./cmd/crun
BIN_DIR := bin

GO := go
GOFLAGS := -trimpath
LDFLAGS := -s -w

.PHONY: all build run clean fmt vet test

all: build

build:
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY) $(CMD_DIR)

run: build
	sudo $(BIN_DIR)/$(BINARY)

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

clean:
	rm -rf $(BIN_DIR)
