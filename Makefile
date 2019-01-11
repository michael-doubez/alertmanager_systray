# Copyright 2019 The alertmanager_systray authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

all: format build check

# ---- Main targets
.PHONY: all format build check clean

# all packages
PKGS=./...
GOBIN=$(shell pwd)/bin
GOFILES=alertmanager_systray.go alertmanager.go config.go

# tools and path
GO111MODULE?=on
GO=GO111MODULE=$(GO111MODULE) go
GOFMT=$(GO)fmt
GOFLAGS=-mod=vendor

format:
	@echo "Formating code"
	@$(GO) fmt $(PKGS)

build:
	@echo "Build alertmanager_systray"
	@$(GO) build $(GOFLAGS) -ldflags -H=windowsgui $(GOFILES)

check: checkstyle checkgo test

clean:
	@$(GO) clean
	@$(GO) mod tidy

run:
	@$(GO) run $(GOFLAGS) $(GOFILES)

# ---- Checks
.PHONY: checkstyle checkgo test

checkstyle:
	@echo "Check style"
	@$(GOFMT) -d $$(find . -path ./vendor -prune -o -name '*.go')

checkgo:
	@echo "Run vet"
	@$(GO) vet $(PKGS)

test:
	@echo "Run tests"
	@$(GO) test $(GOFLAGS) -v $(PKGS)


