GO ?= go

.PHONY: test # Run the unit tests
test:
	$(GO) test ./... -count=1

.PHONY: cover # Run test coverage
cover: $(shell find . -name \*.go)
	$(GO) test -v -coverprofile=cover.out -covermode=count ./...
	@$(GO) tool cover -html=cover.out

.PHONY: lint # Lint the project
lint: .golangci.yaml
	@pre-commit run --show-diff-on-failure --color=always --all-files

.PHONY: clean # Remove generated files
clean:
	$(GO) clean
	@rm -rf cover.out _build

# WASM
.PHONY: wasm # Build a simple app with Go and TinyGo WASM compilation.
wasm: _build/go.wasm _build/tinygo.wasm

_build/go.wasm: internal/wasm/wasm.go
	@mkdir -p $(@D)
	GOOS=js GOARCH=wasm $(GO) build -o $@ $<

_build/tinygo.wasm: internal/wasm/wasm.go
	@mkdir -p $(@D)
	GOOS=js GOARCH=wasm tinygo build -no-debug -size short -o $@ $<

############################################################################
# Utilities.
.PHONY: brew-lint-depends # Install linting tools from Homebrew
brew-lint-depends:
	brew install golangci-lint

.PHONY: debian-lint-depends # Install linting tools on Debian
debian-lint-depends:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sudo sh -s -- -b /usr/bin v1.62.2

.PHONY: install-generators # Install Go code generators
install-generators:
	@$(GO) install golang.org/x/tools/cmd/stringer@v0.24.0

.PHONY: generate # Generate Go code
generate:
	@$(GO) generate ./...

## .git/hooks/pre-commit: Install the pre-commit hook
.git/hooks/pre-commit:
	@printf "#!/bin/sh\nmake lint\n" > $@
	@chmod +x $@

submodules:
	git submodule update --init

update-submodules:
	git pull --recurse-submodules
