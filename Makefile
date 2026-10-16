.PHONY: dev build test test-integration fmt fmt-check vet check tools appimage bundle bundle-local clean

GO ?= go
PYTHON ?= python3
BACKEND ?= wayland,x11

# Prefer native Wayland on Linux and fall back to X11.
dev:
	GDK_BACKEND=$(BACKEND) $(GO) run -buildvcs=false .

build:
	$(GO) build -buildvcs=false -trimpath -ldflags='-s -w' -o build/curb .

test:
	$(GO) test -buildvcs=false -race ./...

test-integration:
	CURB_INTEGRATION=1 $(GO) test -buildvcs=false -race -count=1 ./internal/download

fmt:
	gofmt -w main.go internal

fmt-check:
	@test -z "$$(gofmt -l main.go internal)" || { gofmt -l main.go internal; exit 1; }

vet:
	$(GO) vet -buildvcs=false ./...

check: fmt-check vet test
	$(PYTHON) -m compileall -q scripts
	PYTHONPATH=scripts $(PYTHON) -m unittest discover -s scripts/tests
	sh -n packaging/linux/install.sh
	sh -n packaging/linux/AppRun packaging/linux/build-ffmpeg.sh
	sh -n packaging/native/build-ffmpeg.sh

tools:
	$(PYTHON) scripts/bundle.py --prepare-only

appimage bundle:
	$(PYTHON) scripts/bundle.py

bundle-local:
	$(PYTHON) scripts/bundle.py --skip-download

# Keep downloaded tools and cached upstream archives for the next build.
clean:
	rm -rf build/curb build/mygo dist
