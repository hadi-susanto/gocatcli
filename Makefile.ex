# Fork-only makefile (main-ex branch of hadi-susanto/gocatcli-ex).
# Builds the gocatcli-ex binary: make -f Makefile.ex

GOBIN=$(shell pwd)/bin
BIN=$(GOBIN)/gocatcli-ex
CGO=0
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)
# short sha, "-dirty" suffix on uncommitted changes, tags ignored
GITCOMMIT=$(shell git describe --always --dirty --abbrev=7 --exclude='*')
SRC="./cmd/gocatcli-ex"
INSTALL_FLAG=-v -ldflags "-s -w -X main.commit=$(GITCOMMIT)"

all: build

build:
	CGO_ENABLED=$(CGO) GO111MODULE=on GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(INSTALL_FLAG) -o $(BIN) $(SRC)

build-linux:
	CGO_ENABLED=$(CGO) GO111MODULE=on GOOS=linux GOARCH=arm go build $(INSTALL_FLAG) -o $(BIN)-linux-arm $(SRC)
	CGO_ENABLED=$(CGO) GO111MODULE=on GOOS=linux GOARCH=arm64 go build $(INSTALL_FLAG) -o $(BIN)-linux-arm64 $(SRC)
	CGO_ENABLED=$(CGO) GO111MODULE=on GOOS=linux GOARCH=386 go build $(INSTALL_FLAG) -o $(BIN)-linux-386 $(SRC)
	CGO_ENABLED=$(CGO) GO111MODULE=on GOOS=linux GOARCH=amd64 go build $(INSTALL_FLAG) -o $(BIN)-linux-amd64 $(SRC)

build-windows:
	CGO_ENABLED=$(CGO) GO111MODULE=on GOOS=windows GOARCH=arm64 go build $(INSTALL_FLAG) -o $(BIN)-windows-arm64.exe $(SRC)
	CGO_ENABLED=$(CGO) GO111MODULE=on GOOS=windows GOARCH=386 go build $(INSTALL_FLAG) -o $(BIN)-windows-386.exe $(SRC)
	CGO_ENABLED=$(CGO) GO111MODULE=on GOOS=windows GOARCH=amd64 go build $(INSTALL_FLAG) -o $(BIN)-windows-amd64.exe $(SRC)

build-darwin:
	CGO_ENABLED=$(CGO) GO111MODULE=on GOOS=darwin GOARCH=amd64 go build $(INSTALL_FLAG) -o $(BIN)-darwin-amd64 $(SRC)
	CGO_ENABLED=$(CGO) GO111MODULE=on GOOS=darwin GOARCH=arm64 go build $(INSTALL_FLAG) -o $(BIN)-darwin-arm64 $(SRC)

build-all: build-linux build-windows build-darwin

clean:
	@rm -f $(BIN) $(BIN)-*

# release tag: v<upstream x.y.z>-ex-YY-MM-DD (".2", ".3"... for extra releases the same day)
# pushing it triggers .github/workflows/release-ex.yml
RELEASE_BRANCH=main-ex
TAG_DATE=$(shell date +%y-%m-%d)
UPSTREAM_VERSION=$(shell sed -n 's/^[[:space:]]*version[[:space:]]*=[[:space:]]*"\([^"]*\)".*/\1/p' internal/commands/root.go)

tag:
	@test -n "$(UPSTREAM_VERSION)" || { echo "cannot read version from internal/commands/root.go"; exit 1; }
	@test "$$(git rev-parse --abbrev-ref HEAD)" = "$(RELEASE_BRANCH)" || { echo "not on $(RELEASE_BRANCH)"; exit 1; }
	@test -z "$$(git status --porcelain)" || { echo "working tree not clean"; exit 1; }
	@existing="$$(git tag --points-at HEAD --list 'v*-ex-*')"; \
	test -z "$$existing" || { echo "HEAD already tagged: $$existing"; exit 1; }
	@base="v$(UPSTREAM_VERSION)-ex-$(TAG_DATE)"; tag="$$base"; n=1; \
	while git rev-parse -q --verify "refs/tags/$$tag" >/dev/null; do n=$$((n+1)); tag="$$base.$$n"; done; \
	git tag -a "$$tag" -m "gocatcli-ex $$tag" && echo "created tag $$tag"

release: tag
	@tag="$$(git tag --points-at HEAD --list 'v*-ex-*')"; \
	git push origin "$(RELEASE_BRANCH)" "$$tag"

.PHONY: build build-linux build-windows build-darwin build-all clean all tag release
