.DEFAULT_GOAL := check

DEPS_STAMP := .make/deps.stamp

.PHONY: check ci deps lint test build demos _lint _test _build _demos

check: deps
	@$(MAKE) _lint & \
	lint_pid="$$!"; \
	$(MAKE) _test & \
	test_pid="$$!"; \
	lint_status=0; \
	test_status=0; \
	wait "$$lint_pid" || lint_status="$$?"; \
	wait "$$test_pid" || test_status="$$?"; \
	if [ "$$lint_status" -ne 0 ] || [ "$$test_status" -ne 0 ]; then \
		exit 1; \
	fi
	@$(MAKE) _build
	@$(MAKE) _demos

ci: deps _lint _test _build _demos

deps: $(DEPS_STAMP)

$(DEPS_STAMP): go.mod go.sum
	@mkdir -p "$(@D)"
	go mod download
	@touch "$@"

lint: deps _lint

test: deps _test

build: deps _build

demos: deps _demos

_lint:
	go fmt ./...
	go vet ./...

_test:
	go test ./...

_build:
	go build ./...

_demos:
	@demo_packages="$$(go list ./cmd/... 2>/dev/null)"; \
	if [ -n "$$demo_packages" ]; then \
		tmp_dir="$$(mktemp -d)"; \
		trap 'rm -rf "$$tmp_dir"' EXIT; \
		for package in $$demo_packages; do \
			name="$$(basename "$$package")"; \
			go build -o "$$tmp_dir/$$name" "$$package"; \
		done; \
	fi
