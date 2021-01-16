GO        ?= go
DEBUG     ?= 0
VERBOSE   ?= 0

ifneq ($(DEBUG),0)
GO_TEST_FLAGS        += -count=1
endif
ifneq ($(VERBOSE),0)
GO_TEST_FLAGS        += -v
GO_TEST_BENCH_FLAGS  += -v
endif

GO_TOOLS_GOLANGCI_LINT ?= $(shell $(GO) env GOPATH)/bin/golangci-lint

# -- test ----------------------------------------------------------------------

.PHONY: test bench
.ONESHELL: test bench lint

test:
	$(GO) test $(GO_TEST_FLAGS) ./...

bench:
	$(GO) test $(GO_TEST_FLAGS) -bench=.* ./...

lint: $(GO_TOOLS_GOLANGCI_LINT)
	$(GO_TOOLS_GOLANGCI_LINT) run

# -- tools ---------------------------------------------------------------------

.PHONY: tools

tools: $(GO_TOOLS_GOLANGCI_LINT)

$(GO_TOOLS_GOLANGCI_LINT):
	GO111MODULE=on $(GO) get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.27.0


# -- $(GO) mod --------------------------------------------------------------------

.PHONY: go-mod-verify go-mod-tidy
.ONESHELL: go-mod-verify go-mod-tidy

go-mod-verify:
	$(GO) mod download
	git diff --quiet go.* || git diff --exit-code go.* || exit 1

go-mod-tidy:
	$(GO) mod download
	$(GO) mod tidy

# -- release -------------------------------------------------------------------

.PHONY: tags

tags:
	PRETTY_DIR=$$(sed -e "s#^./##" -e "s#/\$$##" <<<"$(CURDIR)"); \
	LAST_TAG=$$(git describe --abbrev=0 --match="$$PRETTY_DIR/*" | sed -e "s#^$$PRETTY_DIR/##"); \
	CHANGED_FILES=$$(git diff --name-only --ignore-all-space --ignore-space-change $$PRETTY_DIR/$$LAST_TAG..HEAD -- $$PRETTY_DIR); \
	if [ -z "$$CHANGED_FILES" ]; then \
		echo "$$PRETTY_DIR does not need tagging"; \
		continue; \
	fi; \
	NEXT_TAG=$$(gorelease -base=$$LAST_TAG | grep 'Suggested version' | sed -e 's#Suggested version: .* (with tag \(.*\))#\1#'); \
	git rev-parse $$NEXT_TAG -- > /dev/null 2>&1; \
	if [ "$$?" -gt "0" ]; then \
		if [ -n "$$TAG" ]; then \
			echo "Tagging $$NEXT_TAG"; \
			git tag -s -m $$NEXT_TAG $$NEXT_TAG; \
		else \
			echo "Would tag $$NEXT_TAG"; \
		fi; \
	else \
		echo "No new tag for $$PRETTY_DIR"; \
	fi;
