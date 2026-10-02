# ge
# 2023-04-05
# 2023-06-02
# 2026-10-01

# Build for 64-bit with tags for Linux
# $ make OS=linux BIT=64 TAG="cts cgo release"

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

# Tag (スペース区切りで複数指定可能)
TAG ?= cts cgo debug

# Locale (ja_JP, en_US ...)
LOCALE ?=

# Build tags (collect only tag names)
# Goの -tags にはスペース区切りの文字列をそのまま渡せます
BUILD_TAG_NAMES = $(TAG) $(LOCALE)

# Build output directory
BUILD_DIR = .

# Binary filename
BINARY_NAME = ge

# Build options
BUILD_OPTIONS =

# Gitのコミットハッシュの取得
GIT_COMMIT := $(shell git describe --dirty --always | sed 's/-dirty//' 2>/dev/null || echo "not present")

# ビルド日時の取得
BUILD_TIME := $(shell date +%Y-%m-%dT%H:%M:%S%z)

# Setting build flags based on target
ifeq ($(OS), windows)
	# Build flags for Windows
	BUILD_FLAGS = GOOS=windows GOARCH=$(GOARCH)
	BINARY_NAME := $(BINARY_NAME).exe # To prevent from becoming recursively expanded variable
else ifeq ($(OS), linux)
	# Build flags for Linux
	BUILD_FLAGS = GOOS=linux GOARCH=$(GOARCH)
else
	$(error "Invalid OS specified. Use OS=[linux|windows]")
endif

# TAG 内に特定のキーワードが含まれているかでビルドオプションを切り替える
ifneq ($(filter release,$(TAG)),)
	BUILD_OPTIONS = -ldflags "-s -w -X 'main.buildTime=$(BUILD_TIME)' -X 'main.gitCommit=$(GIT_COMMIT)'" -trimpath -a
else ifneq ($(filter develop,$(TAG)),)
	# develop向けのオプションがあればここに記述
else ifneq ($(filter debug,$(TAG)),)
	BUILD_OPTIONS = -ldflags "-X 'main.buildTime=$(BUILD_Time)' -X 'main.gitCommit=$(GIT_COMMIT)'"
endif

# Build target
.PHONY: build
build:
	@echo "Building $(OS) $(BIT)-bit with tags ($(TAG))..."
	@mkdir -p $(BUILD_DIR)
	$(BUILD_FLAGS) go build -tags "$(BUILD_TAG_NAMES)" $(BUILD_OPTIONS) -o $(BUILD_DIR)/$(BINARY_NAME)

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
	@echo '  TAG="cts cgo debug"'
	@echo clean
	@echo help
