GO ?= go
GOCACHE ?= $(CURDIR)/.cache/go-build
COVERAGE_PROFILE ?= coverage.out
COVERPKG ?= ./...

.PHONY: test coverage

test:
	mkdir -p $(GOCACHE)
	GOCACHE=$(GOCACHE) $(GO) test -cover ./...

coverage:
	mkdir -p $(GOCACHE)
	GOCACHE=$(GOCACHE) $(GO) test -coverpkg=$(COVERPKG) -coverprofile=$(COVERAGE_PROFILE) ./...
	GOCACHE=$(GOCACHE) $(GO) tool cover -func=$(COVERAGE_PROFILE)
