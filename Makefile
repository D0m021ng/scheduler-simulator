# init project path
HOMEDIR := $(shell pwd)
OUTDIR  := $(HOMEDIR)/output

# init command params
GO      := go
GOPATH  := $(shell $(GO) env GOPATH)
GOMOD   := $(GO) mod
GOBUILD := $(GO) build
GOTEST  := $(GO) test -gcflags="-N -l"
GOPKGS  := $$($(GO) list ./...| grep -vE "vendor")
export PATH := $(GOPATH)/bin/:$(PATH)

# test cover files
COVPROF := $(HOMEDIR)/covprof.out  # coverage profile
COVFUNC := $(HOMEDIR)/covfunc.txt  # coverage profile information for each function
COVHTML := $(HOMEDIR)/covhtml.html # HTML representation of coverage profile

GIT_COMMIT  = `git rev-parse HEAD`
GIT_DATE    = `date "+%Y-%m-%d %H:%M:%S"`
GIT_VERSION = `git --version`
GIT_BRANCH  = `git rev-parse --abbrev-ref HEAD`

LD_FLAGS    = " \
    -X 'github.com/D0m021ng/scheduler-simulator/pkg/version.GitVersion=${GIT_VERSION}' \
    -X 'github.com/D0m021ng/scheduler-simulator/pkg/version.GitCommit=${GIT_COMMIT}' \
    -X 'github.com/D0m021ng/scheduler-simulator/pkg/version.GitBranch=${GIT_BRANCH}' \
    -X 'github.com/D0m021ng/scheduler-simulator/pkg/version.BuildDate=${GIT_DATE}' \
    "

# make, make all
all: prepare compile package

# make prepare, download dependencies
prepare: gomod

gomod:
	$(GO) env -w GO111MODULE=on
	$(GO) env -w CGO_ENABLED=0
	$(GOMOD) download

# make compile
compile: build

build:
	$(GOBUILD) -ldflags ${LD_FLAGS} -trimpath -o $(HOMEDIR)/simctl $(HOMEDIR)/cmd/simctl/main.go

# make test, test your code
test: prepare test-case
test-case:
	$(GOTEST) -v -cover $(GOPKGS)

# make package
package:
	mkdir -p $(OUTDIR)/bin
	mv $(HOMEDIR)/simctl   $(OUTDIR)/bin

# make clean
clean:
	$(GO) clean
	rm -rf $(OUTDIR)

# avoid filename conflict and speed up build
.PHONY: all prepare compile test package clean build