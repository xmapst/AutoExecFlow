PACKAGE_NAME          := github.com/xmapst/AutoExecFlow
GOLANG_CROSS_VERSION  ?= latest

.PHONY: all
all: binary copy-binary
	@sha256sum bin/AutoExecFlow* > bin/latest.sha256sum

.PHONY: dev
dev: # generate
	@CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o bin/AutoExecFlow cmd/main.go

swag:
	@swag init --exclude pkg --parseDependencyLevel 3 --dir internal/server/api --outputTypes json -g router.go

# Run code generation
generate:
	@echo "Tidying up Go modules..."
	@go mod tidy
	@echo "Running go generate..."
	@go generate ./...

.PHONY: binary
binary:
	@echo "Building the binary..."
	@rm -fr $(CURDIR)/dist
	@docker run \
		--rm \
		--privileged \
		--network host \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(CURDIR):/go/src/$(PACKAGE_NAME) \
		-w /go/src/$(PACKAGE_NAME) \
		ghcr.io/goreleaser/goreleaser-cross:${GOLANG_CROSS_VERSION} \
		release --clean --auto-snapshot --skip=chocolatey,docker,homebrew,publish,scoop,validate,winget

.PHONY: copy-binary
copy-binary:
	@echo "Copying binaries..."
	@rm -fr $(CURDIR)/bin
	@mkdir -p $(CURDIR)/bin
	@find $(CURDIR)/dist/AutoExecFlow* -type f -not -path "*checksums*" -exec bash -c 'cp -f {} $(CURDIR)/bin/`echo {}|sed "s|$(CURDIR)/dist/||g"|sed "s|/AutoExecFlow||g"`' \;
	@rm -fr dist