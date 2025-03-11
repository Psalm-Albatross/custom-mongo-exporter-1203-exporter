# Define the output directory
OUTPUT_DIR := bin

# Define the platforms to build for
PLATFORMS := linux/386 linux/amd64 darwin/386 darwin/amd64 windows/386 windows/amd64

# Default target
.PHONY: all
all: build

# Build target
.PHONY: build
build: $(PLATFORMS)
	@echo "Build completed. Binaries are in the $(OUTPUT_DIR) directory."

# Build for each platform
$(PLATFORMS):
	@OS=$(shell echo $@ | cut -d'/' -f1); \
	ARCH=$(shell echo $@ | cut -d'/' -f2); \
	OUTPUT_NAME="custom-1203-mongo-exporter-$$OS-$$ARCH"; \
	if [ $$OS = "windows" ]; then \
		OUTPUT_NAME+=".exe"; \
	fi; \
	echo "Building for $$OS/$$ARCH..."; \
	GOOS=$$OS GOARCH=$$ARCH go build -ldflags="-X main.version=v0.0.1-alpha" -o $(OUTPUT_DIR)/$$OUTPUT_NAME ./custom-mongo-exporter.go

# Clean target
.PHONY: clean
clean:
	rm -rf $(OUTPUT_DIR)
	@echo "Clean completed. $(OUTPUT_DIR) directory removed."
