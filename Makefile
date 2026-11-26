GOBIN := $(shell go env GOBIN)
ifeq ($(GOBIN),)
GOBIN := $(shell go env GOPATH)/bin
endif

.PHONY: build test install uninstall check clean help
.DEFAULT_GOAL := help

help:
	@echo "Targets:"
	@echo "  build       Build ./ispec in the project root"
	@echo "  test        Run the Go test suite"
	@echo "  install     go install the binary, then install the agent harness integration"
	@echo "  check       Report whether the installed harness integration is current"
	@echo "  uninstall   Remove the harness integration (leaves the binary)"
	@echo "  clean       Remove the locally built binary"
	@echo
	@echo "Installs go to $(GOBIN). Nothing in this checkout is referenced"
	@echo "afterwards: the skills are embedded in the binary."

build:
	go build -o ispec ./cmd/ispec

test:
	go test ./...

# The binary installs itself: `go install` puts it on GOBIN, and the copy
# there is what registers the skills and the session-start hook, so the hook
# records an absolute path that does not depend on this checkout surviving.
install:
	go install ./cmd/ispec
	"$(GOBIN)/ispec" harness install

check:
	"$(GOBIN)/ispec" harness check

uninstall:
	"$(GOBIN)/ispec" harness uninstall

clean:
	rm -f ispec
