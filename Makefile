.DEFAULT_GOAL := check

DEPS_STAMP := .make/deps.stamp

.PHONY: check deps lint test build demos

check: lint test build demos

deps: $(DEPS_STAMP)

$(DEPS_STAMP): go.mod go.sum
	@mkdir -p "$(@D)"
	go mod download
	@touch "$@"

lint: deps
	go fmt ./...
	go vet ./...

test: deps
	go test ./...

build: deps
	go build ./...

demos: deps
	@demo_packages="$$(go list ./cmd/... 2>/dev/null)"; \
	if [ -n "$$demo_packages" ]; then \
		tmp_dir="$$(mktemp -d)"; \
		trap 'rm -rf "$$tmp_dir"' EXIT; \
		for package in $$demo_packages; do \
			name="$$(basename "$$package")"; \
			go build -o "$$tmp_dir/$$name" "$$package"; \
		done; \
	fi
