GO ?= go

SRC_DIR := src
DST_DIR := pub
WASM_EXEC := $(shell tinygo env TINYGOROOT)/targets/wasm_exec.js
# WASM_EXEC := $(shell go env GOROOT)/misc/wasm/wasm_exec.js

playground: $(DST_DIR)/play.wasm $(DST_DIR)/index.html $(DST_DIR)/wasm_exec.js $(DST_DIR)/play.css

ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))
$(DST_DIR)/play.wasm: $(SRC_DIR)/main.go
	@mkdir -p $(@D)
	GOOS=js GOARCH=wasm tinygo build -no-debug -size short -o $@ $<
#	cd $(SRC_DIR); GOOS=js GOARCH=wasm go build -o $(ROOT_DIR)/$@ $$(basename "$<")

$(DST_DIR)/play.css: $(SRC_DIR)/play.css
	mkdir -p $(@D)
	cp $< $@

$(DST_DIR)/index.html: $(SRC_DIR)/index.html
	mkdir -p $(@D)
	cp $< $@

$(DST_DIR)/wasm_exec.js: $(WASM_EXEC)
	mkdir -p $(@D)
	cp $< $@

.PHONY: run
run: playground
	python3 -m http.server --directory $(DST_DIR)

.PHONY: brew-lint-depends # Install linting tools from Homebrew
brew-lint-depends:
	brew install golangci-lint

.PHONY: debian-lint-depends # Install linting tools on Debian
debian-lint-depends:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sudo sh -s -- -b /usr/bin v1.59.0

.PHONY: lint # Lint the project
lint: .pre-commit-config.yaml
	@GOOS=js GOARCH=wasm pre-commit run --show-diff-on-failure --color=always --all-files

.PHONY: golangci-lint # Run golangci-lint
golangci-lint: .golangci.yaml
	golangci-lint run --fix --timeout=5m

.PHONY: clean
clean:
	rm -rf $(DST_DIR)
