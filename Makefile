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
	@echo "Building $(NAME) $(VERSION) for $(shell go env GOOS)/$(shell go env GOARCH)..."
	@CGO_ENABLED=0 go build -ldflags "-X github.com/iansinnott/browser-gopher/cmd.Version=$(VERSION)" -o $(OUTDIR)/$(NAME)
	@echo "Done."