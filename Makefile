GO ?= go

SRC_DIR := src
DST_DIR := pub

# explicitly build the playground with _vendor/tinygo until
# https://github.com/tinygo-org/tinygo/issues/4873 fixed.

playground: $(DST_DIR)/play.wasm $(DST_DIR)/index.html $(DST_DIR)/wasm_exec.js $(DST_DIR)/play.css

ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))
$(DST_DIR)/play.wasm: $(SRC_DIR)/main.go _vendor/tinygo/bin/tinygo
	@mkdir -p $(@D)
	GOOS=js GOARCH=wasm ./_vendor/tinygo/bin/tinygo build -no-debug -size short -o $@ $<
#	cd $(SRC_DIR); GOOS=js GOARCH=wasm go build -o $(ROOT_DIR)/$@ $$(basename "$<")

$(DST_DIR)/play.css: $(SRC_DIR)/play.css
	mkdir -p $(@D)
	cp $< $@

$(DST_DIR)/index.html: $(SRC_DIR)/index.html
	mkdir -p $(@D)
	version=$$(grep jsonpath go.mod | awk '{print $$3}'); cat $< | sed -e "s!{{version}}!$${version}!g" > $@

$(DST_DIR)/wasm_exec.js: _vendor/tinygo/bin/tinygo
	mkdir -p $(@D)
	cp $(shell ./_vendor/tinygo/bin/tinygo env TINYGOROOT)/targets/wasm_exec.js $@

.PHONY: run
run: playground
	python3 -m http.server --directory $(DST_DIR)

.PHONY: brew-lint-depends # Install linting tools from Homebrew
brew-lint-depends:
	brew install golangci-lint

.PHONY: debian-lint-depends # Install linting tools on Debian
debian-lint-depends:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sudo sh -s -- -b /usr/bin v2.5.0

.PHONY: lint # Lint the project
lint: .pre-commit-config.yaml
	@GOOS=js GOARCH=wasm pre-commit run --show-diff-on-failure --color=always --all-files

.PHONY: golangci-lint # Run golangci-lint
golangci-lint: .golangci.yaml
	golangci-lint run --fix --timeout=5m

## .git/hooks/pre-commit: Install the pre-commit hook
.git/hooks/pre-commit:
	@printf "#!/bin/sh\nmake lint\n" > $@
	@chmod +x $@

.PHONY: clean
clean:
	rm -rf $(DST_DIR)

_vendor/tinygo/bin/tinygo:
	brew install binaryen
	mkdir -p _vendor
	cd _vendor && curl -L https://github.com/tinygo-org/tinygo/releases/download/v0.36.0/tinygo0.36.0.darwin-arm64.tar.gz | tar zxf -
