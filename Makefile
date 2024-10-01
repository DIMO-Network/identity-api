.PHONY: all deps docker docker-cgo clean docs test test-race fmt lint install deploy-docs

TAGS =

INSTALL_DIR        = $(GOPATH)/bin
DEST_DIR           = ./target
PATHINSTBIN        = $(abspath $(DEST_DIR)/bin)
PATHINSTDOCKER     = $(DEST_DIR)/docker

# Add target bin dir to PATH
export PATH := $(PATHINSTBIN):$(PATH)
OLDSHELL := $(SHELL)
SHELL = env PATH=$(PATH) $(OLDSHELL)

VERSION   := $(shell git describe --tags || echo "v0.0.0")
VER_CUT   := $(shell echo $(VERSION) | cut -c2-)
VER_MAJOR := $(shell echo $(VER_CUT) | cut -f1 -d.)
VER_MINOR := $(shell echo $(VER_CUT) | cut -f2 -d.)
VER_PATCH := $(shell echo $(VER_CUT) | cut -f3 -d.)
VER_RC    := $(shell echo $(VER_PATCH) | cut -f2 -d-)
DATE      := $(shell date +"%Y-%m-%dT%H:%M:%SZ")

LD_FLAGS   =
GO_FLAGS   =
DOCS_FLAGS =
NAME?="new"

GOLANGCI_VERSION   = latest
# Get binary versions from go.mod
GQLGEN_VERSION     =  $(shell go list -m -f '{{.Version}}' github.com/99designs/gqlgen)
GOOSE_VERSION      =  $(shell go list -m -f '{{.Version}}' github.com/pressly/goose/v3)
SQLBOILER_VERSION  =  $(shell go list -m -f '{{.Version}}' github.com/volatiletech/sqlboiler/v4)

APPS = identity-api

help:
	@echo "\nSpecify a subcommand:\n"
	@grep -hE '^[0-9a-zA-Z_-]+:.*?## .*$$' ${MAKEFILE_LIST} | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[0;36m%-20s\033[m %s\n", $$1, $$2}'
	@echo ""

all: $(APPS)

install: $(APPS)
	@mkdir -p bin
	@cp $(PATHINSTBIN)/$(APPS) ./bin/

deps:
	@go mod tidy
	@go mod vendor

SOURCE_FILES = $(shell find graph internal models cmd -type f -name "*.go")


$(PATHINSTBIN)/%: $(SOURCE_FILES) 
	@go build $(GO_FLAGS) -tags "$(TAGS)" -ldflags "$(LD_FLAGS) " -o $@ ./cmd/$*

$(APPS): %: $(PATHINSTBIN)/%

docker-tags:
	@echo "latest,$(VER_CUT),$(VER_MAJOR).$(VER_MINOR),$(VER_MAJOR)" > .tags

docker-rc-tags:
	@echo "latest,$(VER_CUT),$(VER_MAJOR)-$(VER_RC)" > .tags

docker-cgo-tags:
	@echo "latest-cgo,$(VER_CUT)-cgo,$(VER_MAJOR).$(VER_MINOR)-cgo,$(VER_MAJOR)-cgo" > .tags

docker: deps
	@docker build -f ./resources/docker/Dockerfile . -t dimozone/identity-api:$(VER_CUT)
	@docker tag dimozone/identity-api:$(VER_CUT) dimozone/identity-api:latest

docker-cgo: deps
	@docker build -f ./resources/docker/Dockerfile.cgo . -t dimozone/identity-api:$(VER_CUT)-cgo
	@docker tag dimozone/identity-api:$(VER_CUT)-cgo dimozone/identity-api:latest-cgo

fmt:
	@go list -f {{.Dir}} ./... | xargs -I{} gofmt -w -s {}
	@go mod tidy

lint: ## Run linter.
	@golangci-lint run

test: ## Run all package tests.
	@go test $(GO_FLAGS) -timeout 3m -p=1 ./...

clean: ## Remove previous builds.
	rm -rf $(PATHINSTBIN)
	rm -rf $(DEST_DIR)/dist
	rm -rf $(PATHINSTDOCKER)

run: $(APPS) ## Run the app.
	$(PATHINSTBIN)/$(APPS)

migrate: $(APPS) ## Run database migrations.
	$(PATHINSTBIN)/$(APPS) migrate

sql: ## Create a new SQL migration file. Use the NAME variable to set the name: "make sql NAME=dcn_table".
	@goose -version
	goose  -dir migrations -s create $(NAME) sql 

boil: ## Generate SQLBoiler models.
	@sqlboiler --version
	sqlboiler psql --no-tests --wipe

gql: ## Generate gqlgen code.
	@gqlgen version
	gqlgen generate

tools-golangci-lint: ## Install golangci-lint dependency.
	@mkdir -p $(PATHINSTBIN)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(PATHINSTBIN) $(GOLANGCI_VERSION)

tools-gqlgen: ## Install gqlgen dependency.
	@mkdir -p $(PATHINSTBIN)
	GOBIN=$(PATHINSTBIN) go install github.com/99designs/gqlgen@$(GQLGEN_VERSION)

tools-goose: ## Install goose dependency.
	@mkdir -p $(PATHINSTBIN)
	GOBIN=$(PATHINSTBIN) go install github.com/pressly/goose/v3/cmd/goose@$(GOOSE_VERSION)

tools-sqlboiler: ## Install sqlboiler dependency.
	@mkdir -p $(PATHINSTBIN)
	GOBIN=$(PATHINSTBIN) go install github.com/volatiletech/sqlboiler/v4@$(SQLBOILER_VERSION)

tools: tools-golangci-lint tools-gqlgen tools-goose tools-sqlboiler ## Install all tool dependencies.
