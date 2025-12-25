# ge
# 2023-04-05
# 2023-06-02

# make help

# Build for 64-bit with release tag for Linux
# $ make OS=linux BIT=64 TAG=release

# Build for 32-bit with develop tag for Windows
# $ make OS=windows BIT=32 TAG=develop

# Builds will be executed according to the specified options, and the build results will be generated in the ../build directory.
# Additionally, using the make clean command, you can remove the build results.
# Modify the build options and build output directory settings as needed to create a Makefile tailored to your project.
# Please refer to it.

# How to write AND condition:
# ifeq ($(variable1), value1)
#   ifeq ($(variable2), value2)
#     # Processing when both conditions are met
#   endif
# endif

# Build for OS (linux, windows)
OS ?= linux

# Build for Bit (32, 64)
BIT ?= 64
ifeq ($(BIT), 64)
	GOARCH = amd64
else ifeq ($(BIT), 32)
    GOARCH = 386
else
    $(error "Invalid BIT specified. Use BIT=[64|32]")
endif

# Tag (release, develop, debug)
# TAG ?= release
TAG ?= debug

# Build output directory
# BUILD_DIR = ../../build
BUILD_DIR = .

# Binary filename
BINARY_NAME = ge

# Build options
# BUILD_OPTIONS = -ldflags "-s -w"
BUILD_OPTIONS =

# Build tags
BUILD_TAGS ?= -tags debug

# Environment variables
# export CGO_ENABLED=1

# Gitのコミットハッシュの取得
GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "not present")

# ビルド日時の取得
BUILD_TIME := $(shell date +%Y-%m-%d\ %H:%M:%S)

# Setting build flags based on target
ifeq ($(OS), windows)
    # Build flags for Windows
    BUILD_FLAGS = GOOS=windows GOARCH=$(GOARCH)
	# = 	Recursively expanded variable
	# := 	Simply expanded variable
    BINARY_NAME := $(BINARY_NAME).exe # To prevent from becoming recursively expanded variable

	# ifeq ($(BIT), 32)
	# 	export CC=i686-w64-mingw32-gcc
	# else ifeq ($(BIT), 64)
	# 	export CC=x86_64-w64-mingw32-gcc
	# endif

	# ifeq ($(TAG), release)
	# 	BUILD_OPTIONS = -ldflags "-s -w -H windowsgui"
	# endif
	# BUILD_OPTIONS += -installsuffix cgo
else ifeq ($(OS), linux)
    # Build flags for Linux
    BUILD_FLAGS = GOOS=linux GOARCH=$(GOARCH)

	# ifeq ($(TAG), release)
	# 	BUILD_OPTIONS = -ldflags "-s -w"
	# endif
else
    $(error "Invalid OS specified. Use OS=[linux|windows]")
endif

ifeq ($(TAG), release)
	BUILD_OPTIONS = -ldflags "-s -w -X 'main.buildTime=$(BUILD_TIME)' -X 'main.gitCommit=$(GIT_COMMIT)'" -trimpath -a
    BUILD_TAGS += -tags release
else ifeq ($(TAG), develop)
    BUILD_TAGS += -tags develop
else ifeq ($(TAG), debug)
	BUILD_OPTIONS = -ldflags "-X 'main.buildTime=$(BUILD_TIME)' -X 'main.gitCommit=$(GIT_COMMIT)'"
    BUILD_TAGS += -tags debug
else
    $(error "Invalid TAG specified. Use TAG=[release|develop|debug]")
endif

# Build target
.PHONY: build
build:
	@echo "Building $(OS) $(BIT)-bit with $(TAG) tag..."
	@mkdir -p $(BUILD_DIR)
	$(BUILD_FLAGS) go build $(BUILD_TAGS) $(BUILD_OPTIONS) -o $(BUILD_DIR)/$(BINARY_NAME)

# Clean target
.PHONY: clean
clean:
	@echo "Cleaning up..."
	# rm -rf $(BUILD_DIR)
	trash $(BUILD_DIR)

# Help target
.PHONY: help
help:
	@echo build:
	@echo '  OS=[linux|windows]'
	@echo '  BIT=[64|32]'
	@echo '  TAG=[release|develop|debug]'
	@echo clean
	@echo help
