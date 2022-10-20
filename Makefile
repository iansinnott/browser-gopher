all:

build:
	@go build -ldflags "-X github.com/iansinnott/browser-gopher/cmd.Version=$$(git describe --tags)"