.PHONY: help build-mac build-fedora docker-image clean

help:
	@echo "LocalShare Build Automation"
	@echo "==========================="
	@echo "Available targets:"
	@echo "  make build-mac     - Build the native macOS ARM64 app bundle"
	@echo "  make docker-image  - Build the Fedora Linux x86_64 Docker compiler environment"
	@echo "  make build-fedora  - Build the native Fedora Linux x86_64 binary inside Docker"
	@echo "  make clean         - Clean the build bin output directory"

build-mac:
	@echo "Building native macOS ARM64 bundle..."
	wails build

docker-image:
	@echo "Building Fedora AMD64 compiler Docker image..."
	docker build --platform linux/amd64 -f Dockerfile.fedora -t localshare-fedora-builder-amd64 .

build-fedora:
	@echo "Building native Fedora Linux x86_64 ELF binary..."
	docker run --rm -e CI=true -v $(shell pwd):/app localshare-fedora-builder-amd64 wails build -platform linux/amd64 -tags webkit2_41
	@if [ -f build/bin/localshare ]; then \
		mv build/bin/localshare build/bin/localshare-linux-amd64; \
		echo "Fedora AMD64 binary built at: build/bin/localshare-linux-amd64"; \
	else \
		echo "Build failed or binary not found."; \
		exit 1; \
	fi

clean:
	@echo "Cleaning build bin output folder..."
	rm -rf build/bin/localshare*
