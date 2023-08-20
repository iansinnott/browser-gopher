NAME = $(shell basename `pwd`)
VERSION = $(shell git describe --tags --always)
OUTDIR = dist/$(NAME)-$(VERSION)
	
all:
	
dist:
	@mkdir -p dist/$(NAME)-$(VERSION)

.PHONY: outdir
outdir:
	@echo "dist/$(NAME)-$(VERSION)"

build: dist
	@echo "Building with system defaults..."
	@CGO_ENABLED=0 go build -ldflags "-X github.com/iansinnott/browser-gopher/cmd.Version=$(VERSION)"
	@echo "Building ..."
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "-X github.com/iansinnott/browser-gopher/cmd.Version=$(VERSION)" -o $(OUTDIR)/$(NAME)-darwin-arm64
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/iansinnott/browser-gopher/cmd.Version=$(VERSION)" -o $(OUTDIR)/$(NAME)-darwin-amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/iansinnott/browser-gopher/cmd.Version=$(VERSION)" -o $(OUTDIR)/$(NAME)-linux-amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-X github.com/iansinnott/browser-gopher/cmd.Version=$(VERSION)" -o $(OUTDIR)/$(NAME)-linux-arm64
	@echo "Done."