.DEFAULT_GOAL := check

DEMO_PACKAGES := $(shell go list ./cmd/... 2>/dev/null)

.PHONY: check lint test build demos

check: lint test build demos

lint:
	go fmt ./...
	go vet ./...

test:
	go test ./...

build:
	go build ./...

demos:
	@if [ -n "$(DEMO_PACKAGES)" ]; then \
		tmp_dir="$$(mktemp -d)"; \
		trap 'rm -rf "$$tmp_dir"' EXIT; \
		for package in $(DEMO_PACKAGES); do \
			name="$$(basename "$$package")"; \
			go build -o "$$tmp_dir/$$name" "$$package"; \
		done; \
	fi
