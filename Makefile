DIST                 := dist
include versions.mk
REPOSITORY          ?= cdaley-akamai/edgeworkers-cli-demo
EDGEWORKERS_RELEASE_URL ?= https://github.com/$(REPOSITORY)/releases/download/edgeworkers-$(EDGEWORKERS_VERSION)
EDGEKV_RELEASE_URL      ?= https://github.com/$(REPOSITORY)/releases/download/edgekv-$(EDGEKV_VERSION)
DEVSERVER_RELEASE_URL   ?= https://github.com/$(REPOSITORY)/releases/download/devserver-$(DEVSERVER_VERSION)

.PHONY: build build-edgeworkers build-edgekv build-devserver format test lint validate-edgeworkers validate-edgekv validate-devserver clean release release-edgeworkers release-edgekv release-devserver gh-release-edgeworkers gh-release-edgekv gh-release-devserver all

build: build-edgeworkers build-edgekv build-devserver

build-edgeworkers: format test lint
	mkdir -p $(DIST)/edgeworkers
	go build -ldflags "-X main.version=$(EDGEWORKERS_VERSION)" -o $(DIST)/edgeworkers/akamai-edgeworkers ./cmd/edgeworkers

build-edgekv: format test lint
	mkdir -p $(DIST)/edgekv
	go build -ldflags "-X main.version=$(EDGEKV_VERSION)" -o $(DIST)/edgekv/akamai-edgekv ./cmd/edgekv

build-devserver: format test lint
	mkdir -p $(DIST)/devserver
	go build -ldflags "-X main.version=$(DEVSERVER_VERSION)" -o $(DIST)/devserver/edgeworkersdevserver ./cmd/edgeworkersdevserver

validate-edgeworkers: build-edgeworkers
	$(DIST)/edgeworkers/akamai-edgeworkers --help
	npx @modelcontextprotocol/inspector --cli ./$(DIST)/edgeworkers/akamai-edgeworkers mcp --method tools/list --format json
	npx @modelcontextprotocol/inspector --cli ./$(DIST)/edgeworkers/akamai-edgeworkers mcp --method tools/call --tool-name hello_world

validate-edgekv: build-edgekv
	$(DIST)/edgekv/akamai-edgekv --help
	npx @modelcontextprotocol/inspector --cli ./$(DIST)/edgekv/akamai-edgekv mcp --method tools/list --format json
	npx @modelcontextprotocol/inspector --cli ./$(DIST)/edgekv/akamai-edgekv mcp --method tools/call --tool-name hello_world

validate-devserver: build-devserver
	$(DIST)/devserver/edgeworkersdevserver --help

format:
	gofmt -w $(shell rg --files -g '*.go')

test:
	go tool gotestsum

lint:
	go vet ./...

clean:
	rm -rf $(DIST)

# release-devserver is intentionally excluded from `release`: it is versioned and
# released independently of EdgeWorkers/EdgeKV.
release: clean release-edgeworkers release-edgekv

release-edgeworkers:
	go run ./build/package/set-cli-version -cli cli.json -command edgeworkers -version $(EDGEWORKERS_VERSION)
	VERSION=$(EDGEWORKERS_VERSION) REPOSITORY=$(REPOSITORY) RELEASE_URL=$(EDGEWORKERS_RELEASE_URL) NAME=edgeworkers COMMAND_PATH=./cmd/edgeworkers DIST=$(DIST)/edgeworkers/$(EDGEWORKERS_VERSION) bash build/package/build-all.sh

release-edgekv:
	go run ./build/package/set-cli-version -cli cli.json -command edgekv -version $(EDGEKV_VERSION)
	VERSION=$(EDGEKV_VERSION) REPOSITORY=$(REPOSITORY) RELEASE_URL=$(EDGEKV_RELEASE_URL) NAME=edgekv COMMAND_PATH=./cmd/edgekv DIST=$(DIST)/edgekv/$(EDGEKV_VERSION) bash build/package/build-all.sh

release-devserver:
	VERSION=$(DEVSERVER_VERSION) REPOSITORY=$(REPOSITORY) RELEASE_URL=$(DEVSERVER_RELEASE_URL) DIST=$(DIST)/devserver/$(DEVSERVER_VERSION) bash build/package/build-devserver.sh

gh-release-edgeworkers:
	gh release create edgeworkers-$(EDGEWORKERS_VERSION) $(DIST)/edgeworkers/$(EDGEWORKERS_VERSION)/* --repo $(REPOSITORY) --title "EdgeWorkers CLI $(EDGEWORKERS_VERSION)" --generate-notes

gh-release-edgekv:
	gh release create edgekv-$(EDGEKV_VERSION) $(DIST)/edgekv/$(EDGEKV_VERSION)/* --repo $(REPOSITORY) --title "EdgeKV CLI $(EDGEKV_VERSION)" --generate-notes

gh-release-devserver:
	gh release create devserver-$(DEVSERVER_VERSION) $(DIST)/devserver/$(DEVSERVER_VERSION)/* --repo $(REPOSITORY) --title "EdgeWorkersDevServer $(DEVSERVER_VERSION)" --generate-notes

all: test build
