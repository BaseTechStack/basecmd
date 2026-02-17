# Build variables
BINARY_NAME=base
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null | sed 's/^v//' || echo "dev")
COMMIT_HASH=$(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GO_VERSION=$(shell go version | cut -d ' ' -f 3)

# Detect OS and architecture
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

# Set GOOS
ifeq ($(UNAME_S),Darwin)
    GOOS=darwin
else ifeq ($(UNAME_S),Linux)
    GOOS=linux
else ifneq ($(findstring MINGW,$(UNAME_S)),)
    GOOS=windows
    BINARY_NAME=base.exe
else ifneq ($(findstring MSYS,$(UNAME_S)),)
    GOOS=windows
    BINARY_NAME=base.exe
else
    $(error Unsupported operating system: $(UNAME_S))
endif

# Set GOARCH
ifeq ($(UNAME_M),x86_64)
    GOARCH=amd64
else ifeq ($(UNAME_M),amd64)
    GOARCH=amd64
else ifeq ($(UNAME_M),arm64)
    GOARCH=arm64
else ifeq ($(UNAME_M),aarch64)
    GOARCH=arm64
else
    $(error Unsupported architecture: $(UNAME_M))
endif

# Cross-compilation targets
.PHONY: build-all
build-all: clean
	@echo "Building for all platforms..."
	@echo "Building darwin/amd64..."
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o dist/base_darwin_amd64
	@echo "Building darwin/arm64..."
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o dist/base_darwin_arm64
	@echo "Building linux/amd64..."
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o dist/base_linux_amd64
	@echo "Building linux/arm64..."
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o dist/base_linux_arm64
	@echo "Building windows/amd64..."
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o dist/base_windows_amd64.exe

# LDFLAGS for version information
LDFLAGS=-ldflags "-X github.com/base-go/cmd/version.Version=${VERSION} \
                  -X github.com/base-go/cmd/version.CommitHash=${COMMIT_HASH} \
                  -X github.com/base-go/cmd/version.BuildDate=${BUILD_DATE} \
                  -X github.com/base-go/cmd/version.GoVersion=${GO_VERSION}"

.PHONY: all build clean install test dev

all: clean build

build:
	@echo "Building Base CLI..."
	@echo "Version: ${VERSION}"
	@echo "Commit: ${COMMIT_HASH}"
	@echo "Build Date: ${BUILD_DATE}"
	@echo "Go Version: ${GO_VERSION}"
	@echo "OS: ${GOOS}"
	@echo "Architecture: ${GOARCH}"
	@echo "Build command: GOOS=${GOOS} GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BINARY_NAME}"
	GOOS=${GOOS} GOARCH=${GOARCH} go build -v ${LDFLAGS} -o ${BINARY_NAME}

clean:
	@echo "Cleaning..."
	rm -f ${BINARY_NAME}

install: build
	@echo "Installing..."
	mkdir -p ${HOME}/.base
	mv ${BINARY_NAME} ${HOME}/.base/${BINARY_NAME}
	echo "darude" | sudo -S ln -sf ${HOME}/.base/${BINARY_NAME} /usr/local/bin/${BINARY_NAME}

dev: clean
	@echo "Building for development..."
	go build -o ${BINARY_NAME}
	./${BINARY_NAME} version

test:
	@echo "Running tests..."
	go test -v ./...

# Release targets
.PHONY: release release-patch release-minor release-major

CURRENT_VERSION=$(shell git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//' || echo "0.0.0")

release-patch:
	@echo "Creating patch release..."
	$(eval NEW_VERSION=$(shell echo ${CURRENT_VERSION} | awk -F. '{$$NF = $$NF + 1;} 1' | sed 's/ /./g'))
	@git tag -a v${NEW_VERSION} -m "Release v${NEW_VERSION}"
	@git push origin v${NEW_VERSION}
	@echo "Released v${NEW_VERSION}"

release-minor:
	@echo "Creating minor release..."
	$(eval NEW_VERSION=$(shell echo ${CURRENT_VERSION} | awk -F. '{$$(NF-1) = $$(NF-1) + 1; $$NF = 0;} 1' | sed 's/ /./g'))
	@git tag -a v${NEW_VERSION} -m "Release v${NEW_VERSION}"
	@git push origin v${NEW_VERSION}
	@echo "Released v${NEW_VERSION}"

release-major:
	@echo "Creating major release..."
	$(eval NEW_VERSION=$(shell echo ${CURRENT_VERSION} | awk -F. '{$$1 = substr($$1,2) + 1; $$(NF-1) = 0; $$NF = 0;} 1' | sed 's/ /./g'))
	@git tag -a v${NEW_VERSION} -m "Release v${NEW_VERSION}"
	@git push origin v${NEW_VERSION}
	@echo "Released v${NEW_VERSION}"

# Development helpers
.PHONY: fmt lint

fmt:
	@echo "Formatting code..."
	go fmt ./...
	find . -name "*.go" -exec goimports -w {} \;

lint:
	@echo "Linting code..."
	go vet ./...
	golint ./...
