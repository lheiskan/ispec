ISPEC_ROOT := $(abspath $(dir $(firstword $(MAKEFILE_LIST))))
SKILLS := bootstrap defer retrofit session-close spec-session
SKILLS_DST ?= $(HOME)/.claude/skills
BIN_DST ?= $(HOME)/.local/bin

.PHONY: build install uninstall clean help
.DEFAULT_GOAL := help

help:
	@echo "Targets:"
	@echo "  build       Build the ispec-lint binary"
	@echo "  install     Build, then symlink skills and binary into user-global locations"
	@echo "  uninstall   Remove ispec's symlinks (does not delete source or build artifacts)"
	@echo "  clean       Remove the built binary"
	@echo
	@echo "Configurable variables (set on the command line, e.g. make install BIN_DST=...):"
	@echo "  SKILLS_DST  default $(HOME)/.claude/skills"
	@echo "  BIN_DST     default $(HOME)/.local/bin"

build: ispec-lint

ispec-lint: cmd/ispec-lint/main.go go.mod go.sum
	go build ./cmd/ispec-lint

install: build
	@mkdir -p "$(SKILLS_DST)" "$(BIN_DST)"
	@for skill in $(SKILLS); do \
		src="$(ISPEC_ROOT)/.claude/skills/$$skill"; \
		dst="$(SKILLS_DST)/$$skill"; \
		if [ -L "$$dst" ] || [ -e "$$dst" ]; then \
			echo "skipping $$dst (already exists)"; \
		else \
			ln -s "$$src" "$$dst"; \
			echo "linked $$dst -> $$src"; \
		fi; \
	done
	@bin_src="$(ISPEC_ROOT)/ispec-lint"; \
	bin_link="$(BIN_DST)/ispec-lint"; \
	if [ -L "$$bin_link" ] || [ -e "$$bin_link" ]; then \
		echo "skipping $$bin_link (already exists)"; \
	else \
		ln -s "$$bin_src" "$$bin_link"; \
		echo "linked $$bin_link -> $$bin_src"; \
	fi
	@echo
	@echo "Install complete."
	@echo "  skills  -> $(SKILLS_DST)"
	@echo "  binary  -> $(BIN_DST)/ispec-lint"
	@echo "  Make sure $(BIN_DST) is on your \$$PATH."

uninstall:
	@for skill in $(SKILLS); do \
		dst="$(SKILLS_DST)/$$skill"; \
		if [ -L "$$dst" ]; then \
			rm "$$dst"; \
			echo "removed $$dst"; \
		fi; \
	done
	@bin_link="$(BIN_DST)/ispec-lint"; \
	if [ -L "$$bin_link" ]; then \
		rm "$$bin_link"; \
		echo "removed $$bin_link"; \
	fi

clean:
	rm -f ispec-lint
