VERSION := $(shell echo $(shell git describe --tags || git branch --show-current) | sed 's/^v//')
GO_BIN := $(shell echo $(shell which go || echo "/usr/local/go/bin/go" ))

###############################################################################
###                                Build flags                              ###
###############################################################################

LD_FLAGS = -X github.com/bcdevtools/chain-registry-validation-tool/constants.VERSION=$(VERSION)

BUILD_FLAGS := -ldflags '$(LD_FLAGS)'

###############################################################################
###                                  Build                                  ###
###############################################################################

build: go.sum
	@echo "Building chain-registry validation tool..."
	@echo "Flags $(BUILD_FLAGS)"
	@go build -mod=readonly $(BUILD_FLAGS) -o build/crv ./cmd/crv
	@echo "Builded successfully"
.PHONY: build

###############################################################################
###                                 Install                                 ###
###############################################################################

install: go.sum
	@echo "Build flags: $(BUILD_FLAGS)"
	@echo "Installing chain-registry validation tool..."
	@$(GO_BIN) install -mod=readonly $(BUILD_FLAGS) ./cmd/crv
	@echo "Installed successfully"
.PHONY: install