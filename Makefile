BINARY   := athenzctl
PKG      := github.com/fsul7o/athenzctl
CMD      := ./cmd/athenzctl
VERSION  ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE     ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS  := -s -w \
	-X $(PKG)/internal/version.Version=$(VERSION) \
	-X $(PKG)/internal/version.Commit=$(COMMIT) \
	-X $(PKG)/internal/version.Date=$(DATE)

ATHENZ_DIST_DIR    := .local/athenz-distribution
ATHENZ_DIST_REPO   := https://github.com/ctyano/athenz-distribution.git
ATHENZ_DIST_PATCHES := $(wildcard $(PWD)/scripts/patches/athenz-distribution-*.patch)
E2E_CONFIG         := $(PWD)/.local/e2e/config.yaml
E2E_COVERAGE       ?= 1

.PHONY: build test lint tidy install clean snapshot \
	e2e-clone e2e-up e2e-down e2e e2e-focus e2e-sweep e2e-clean

build:
	go build -trimpath -ldflags '$(LDFLAGS)' -o bin/$(BINARY) $(CMD)

test:
	go test ./...

lint:
	go vet ./...

tidy:
	go mod tidy

install:
	go install -trimpath -ldflags '$(LDFLAGS)' $(CMD)

clean:
	rm -rf bin dist

snapshot:
	goreleaser release --snapshot --clean

# Vendored athenz-distribution is treated as disposable (see e2e-clean): any
# locally applied patches are reverted before pulling so `--ff-only` always
# succeeds, then re-applied on top of the fresh checkout every time.
e2e-clone:
	@if [ ! -d $(ATHENZ_DIST_DIR)/.git ]; then \
		mkdir -p $(dir $(ATHENZ_DIST_DIR)); \
		git clone --depth 1 $(ATHENZ_DIST_REPO) $(ATHENZ_DIST_DIR); \
	else \
		git -C $(ATHENZ_DIST_DIR) checkout -- .; \
		git -C $(ATHENZ_DIST_DIR) pull --ff-only; \
	fi
	@for p in $(ATHENZ_DIST_PATCHES); do \
		echo "applying $$p"; \
		git -C $(ATHENZ_DIST_DIR) apply "$$p" || exit 1; \
	done

e2e-up: e2e-clone
	$(MAKE) -C $(ATHENZ_DIST_DIR) \
		deploy-kubernetes-in-docker load-docker-images load-kubernetes-images \
		deploy-kubernetes-crypki-softhsm use-kubernetes-crypki-softhsm \
		deploy-kubernetes-athenz check-kubernetes-athenz deploy-kubernetes-athenz-oauth2
	./scripts/e2e-bootstrap.sh $(ATHENZ_DIST_DIR) $(PWD)/.local/e2e

e2e-down:
	-./scripts/e2e-teardown.sh $(PWD)/.local/e2e
	-$(MAKE) -C $(ATHENZ_DIST_DIR) clean-kubernetes-athenz

e2e:
	ATHENZCTL_E2E_CONFIG=$(E2E_CONFIG) \
	ATHENZCTL_E2E_COVERAGE=$(E2E_COVERAGE) \
	GODOG_TAGS=$${GODOG_TAGS:-~@skip} \
	go test -tags=e2e -count=1 -v ./test/e2e/...

e2e-focus:
	GODOG_TAGS=$${GODOG_TAGS:-@focus} E2E_COVERAGE=0 $(MAKE) e2e

e2e-sweep:
	ATHENZCTL_E2E_CONFIG=$(E2E_CONFIG) \
	GODOG_TAGS=@__sweep_only__ \
	go test -tags=e2e -count=1 -v ./test/e2e/...

e2e-clean:
	rm -rf .local
