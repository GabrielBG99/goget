
BUILDPATH=$(CURDIR)/bin
GO=$(shell which go)
VERSION=$(shell git describe --tags $(shell git rev-list --tags --max-count=1))
LDFLAGS="-X 'main.Version=$(VERSION)'"

EXENAME=goget


install: build ## Install goget to $PATH
	@mv -f $(BUILDPATH)/$(EXENAME) /usr/local/bin
	@$(MAKE) clean

clean: ## Removes the build folder and all its content
	@rm -rf $(BUILDPATH)

build: ## Create the goget's executable
	@if [ ! -d $(BUILDPATH) ] ; then mkdir -p $(BUILDPATH) ; fi
	@$(GO) build -ldflags=$(LDFLAGS) -v -o $(BUILDPATH)/$(EXENAME) main.go

help: ## Display available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
