<<<<<<< HEAD
SHELL=/bin/bash
GIT_URL := "https://github.com/xmapst/osreapi.git"
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
GIT_COMMIT := $(shell git rev-parse HEAD)
VERSION := $(shell git describe --tags)
USER_NAME := $(shell git config user.name)
USER_EMAIL := $(shell git config user.email)
BUILD_TIME := $(shell date +"%Y-%m-%d %H:%M:%S %Z")
LDFLAGS := "-w -s \
-X 'github.com/xmapst/osreapi.Version=$(VERSION)' \
-X 'github.com/xmapst/osreapi.GitUrl=$(GIT_URL)' \
-X 'github.com/xmapst/osreapi.GitBranch=$(GIT_BRANCH)' \
-X 'github.com/xmapst/osreapi.GitCommit=$(GIT_COMMIT)' \
-X 'github.com/xmapst/osreapi.BuildTime=$(BUILD_TIME)' \
-X 'github.com/xmapst/osreapi.UserName=$(USER_NAME)' \
-X 'github.com/xmapst/osreapi.UserEmail=$(USER_EMAIL)' \
"

all: vet fmt windows linux darwin freebsd netbsd openbsd

fmt:
	go fmt ./...

vet:
	go vet ./...

swag:
	swag init -g cmd/osreapi.go -o internal/docs

windows:
	go mod tidy
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -ldflags $(LDFLAGS) -o bin/windows-remote_executor-386.exe cmd/osreapi.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags $(LDFLAGS) -o bin/windows-remote_executor-amd64.exe cmd/osreapi.go

linux:
	go mod tidy
	CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags $(LDFLAGS) -o bin/linux-remote_executor-386 cmd/osreapi.go
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags $(LDFLAGS) -o bin/linux-remote_executor-amd64 cmd/osreapi.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -ldflags $(LDFLAGS) -o bin/linux-remote_executor-arm cmd/osreapi.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags $(LDFLAGS) -o bin/linux-remote_executor-arm64 cmd/osreapi.go
	CGO_ENABLED=0 GOOS=linux GOARCH=ppc64 go build -ldflags $(LDFLAGS) -o bin/linux-remote_executor-ppc64 cmd/osreapi.go
	CGO_ENABLED=0 GOOS=linux GOARCH=ppc64le go build -ldflags $(LDFLAGS) -o bin/linux-remote_executor-ppc64le cmd/osreapi.go
	CGO_ENABLED=0 GOOS=linux GOARCH=mips go build -ldflags $(LDFLAGS) -o bin/linux-remote_executor-mips cmd/osreapi.go
	CGO_ENABLED=0 GOOS=linux GOARCH=mipsle go build -ldflags $(LDFLAGS) -o bin/linux-remote_executor-mipsle cmd/osreapi.go
	CGO_ENABLED=0 GOOS=linux GOARCH=mips64 go build -ldflags $(LDFLAGS) -o bin/linux-remote_executor-mips64 cmd/osreapi.go
	CGO_ENABLED=0 GOOS=linux GOARCH=mips64le go build -ldflags $(LDFLAGS) -o bin/linux-remote_executor-mips64le cmd/osreapi.go

darwin:
	go mod tidy
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags $(LDFLAGS) -o bin/darwin-remote_executor-amd64 cmd/osreapi.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags $(LDFLAGS) -o bin/darwin-remote_executor-arm64 cmd/osreapi.go

freebsd:
	go mod tidy
	CGO_ENABLED=0 GOOS=freebsd GOARCH=386 go build -ldflags $(LDFLAGS) -o bin/freebsd-remote_executor-386 cmd/osreapi.go
	CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -ldflags $(LDFLAGS) -o bin/freebsd-remote_executor-amd64 cmd/osreapi.go
	CGO_ENABLED=0 GOOS=freebsd GOARCH=arm go build -ldflags $(LDFLAGS) -o bin/freebsd-remote_executor-arm cmd/osreapi.go

netbsd:
	go mod tidy
	CGO_ENABLED=0 GOOS=netbsd GOARCH=386 go build -ldflags $(LDFLAGS) -o bin/netbsd-remote_executor-386 cmd/osreapi.go
	CGO_ENABLED=0 GOOS=netbsd GOARCH=amd64 go build -ldflags $(LDFLAGS) -o bin/netbsd-remote_executor-amd64 cmd/osreapi.go
	CGO_ENABLED=0 GOOS=netbsd GOARCH=arm go build -ldflags $(LDFLAGS) -o bin/netbsd-remote_executor-arm cmd/osreapi.go

openbsd:
	go mod tidy
	CGO_ENABLED=0 GOOS=openbsd GOARCH=386 go build -ldflags $(LDFLAGS) -o bin/openbsd-remote_executor-386 cmd/osreapi.go
	CGO_ENABLED=0 GOOS=openbsd GOARCH=amd64 go build -ldflags $(LDFLAGS) -o bin/openbsd-remote_executor-amd64 cmd/osreapi.go
	CGO_ENABLED=0 GOOS=openbsd GOARCH=arm go build -ldflags $(LDFLAGS) -o bin/openbsd-remote_executor-arm cmd/osreapi.go
=======
SHELL=/bin/bash
GIT_URL := "https://github.com/xmapst/osreapi.git"
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD || echo "Unknown")
GIT_COMMIT := $(shell git rev-parse HEAD || echo "Unknown")
VERSION := $(shell git describe --tags || echo "Unknown")
USER_NAME := $(shell git config user.name || echo "Unknown")
USER_EMAIL := $(shell git config user.email || echo "Unknown")
BUILD_TIME := $(shell date +"%Y-%m-%d %H:%M:%S %Z" || echo "Unknown")
CGO_ENABLED := 0
LDFLAGS := "-w -s \
-X 'github.com/xmapst/osreapi/pkg/info.Version=$(VERSION)' \
-X 'github.com/xmapst/osreapi/pkg/info.GitUrl=$(GIT_URL)' \
-X 'github.com/xmapst/osreapi/pkg/info.GitBranch=$(GIT_BRANCH)' \
-X 'github.com/xmapst/osreapi/pkg/info.GitCommit=$(GIT_COMMIT)' \
-X 'github.com/xmapst/osreapi/pkg/info.BuildTime=$(BUILD_TIME)' \
-X 'github.com/xmapst/osreapi/pkg/info.UserName=$(USER_NAME)' \
-X 'github.com/xmapst/osreapi/pkg/info.UserEmail=$(USER_EMAIL)' \
"

all: vet fmt windows linux darwin sha256sum

sha256sum:
	sha256sum bin/*remote_executor* > bin/latest.sha256sum

fmt:
	go fmt ./...

vet:
	go vet ./...

dev:
	go mod tidy
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 go build -trimpath -ldflags $(LDFLAGS) -o bin/linux-remote_executor-amd64 cmd/osreapi.go
	CGO_ENABLED=$(CGO_ENABLED) GOOS=windows GOARCH=amd64 go build -trimpath -ldflags $(LDFLAGS) -o bin/windows-remote_executor-amd64.exe cmd/osreapi.go

windows:
	go mod tidy
	CGO_ENABLED=$(CGO_ENABLED) GOOS=windows GOARCH=amd64 go build -trimpath -ldflags $(LDFLAGS) -o bin/windows-remote_executor-amd64.exe cmd/osreapi.go
	CGO_ENABLED=$(CGO_ENABLED) GOOS=windows GOARCH=arm64 go build -trimpath -ldflags $(LDFLAGS) -o bin/windows-remote_executor-arm64.exe cmd/osreapi.go

linux:
	go mod tidy
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 go build -trimpath -ldflags $(LDFLAGS) -o bin/linux-remote_executor-amd64 cmd/osreapi.go
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=arm64 go build -trimpath -ldflags $(LDFLAGS) -o bin/linux-remote_executor-arm64 cmd/osreapi.go

darwin:
	go mod tidy
	CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags $(LDFLAGS) -o bin/darwin-remote_executor-amd64 cmd/osreapi.go
	CGO_ENABLED=$(CGO_ENABLED) GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags $(LDFLAGS) -o bin/darwin-remote_executor-arm64 cmd/osreapi.go
>>>>>>> githubB
