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
	@CGO_ENABLED=1 go build -ldflags "-X github.com/iansinnott/browser-gopher/cmd.Version=$(VERSION)"
	@echo "Building $(NAME) $(VERSION) for $(shell go env GOOS)/$(shell go env GOARCH)..."
	@CGO_ENABLED=1 go build -ldflags "-X github.com/iansinnott/browser-gopher/cmd.Version=$(VERSION)" -o $(OUTDIR)/$(NAME)
	@echo "Building $(NAME) $(VERSION) for linux/amd64..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags "-X github.com/iansinnott/browser-gopher/cmd.Version=$(VERSION)" -o $(OUTDIR)/$(NAME)-linux-amd64
	@echo "Done."