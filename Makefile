GO_MATRIX += darwin/amd64
GO_MATRIX += linux/amd64
GO_MATRIX += windows/amd64

APP_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_HASH ?= $(shell git show -s --format=%h)

GO_DEBUG_ARGS   ?= -v -ldflags "-X main.version=$(GO_APP_VERSION)+debug -X main.commit=$(GIT_HASH) -X main.date=$(APP_DATE) -X main.builtBy=makefiles"
GO_RELEASE_ARGS ?= -v -ldflags "-X main.version=$(GO_APP_VERSION) -X main.commit=$(GIT_HASH) -X main.date=$(APP_DATE) -X main.builtBy=makefiles -s -w"

-include .makefiles/Makefile
-include .makefiles/pkg/protobuf/v2/Makefile
-include .makefiles/pkg/go/v1/Makefile
-include .makefiles/ext/na4ma4/lib/golangci-lint/v1/Makefile
-include .makefiles/ext/na4ma4/lib/goreleaser/v1/Makefile

.makefiles/ext/na4ma4/%: .makefiles/Makefile
	@curl -sfL https://raw.githubusercontent.com/na4ma4/makefiles-ext/main/v1/install | bash /dev/stdin "$@"

.makefiles/%:
	@curl -sfL https://makefiles.dev/v1 | bash /dev/stdin "$@"

.PHONY: run
run: artifacts/build/debug/$(GOHOSTOS)/$(GOHOSTARCH)/ghtool
	"$<" $(RUN_ARGS)

.PHONY: install
install: $(REQ) $(_SRC) | $(USE)
	$(eval PARTS := $(subst /, ,$*))
	$(eval BUILD := $(word 1,$(PARTS)))
	$(eval OS    := $(word 2,$(PARTS)))
	$(eval ARCH  := $(word 3,$(PARTS)))
	$(eval BIN   := $(word 4,$(PARTS)))
	$(eval ARGS  := $(if $(findstring debug,$(BUILD)),$(DEBUG_ARGS),$(RELEASE_ARGS)))

	CGO_ENABLED=$(CGO_ENABLED) GOOS="$(OS)" GOARCH="$(ARCH)" go install $(ARGS) "./cmd/..."


######################
# Linting
######################

ci:: lint
