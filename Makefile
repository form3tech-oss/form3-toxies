PLATFORM                 := $(shell uname)
NAME                     := $(shell basename $(CURDIR))
GOFMT_FILES              ?= $$(find ./ -name '*.go' | grep -v vendor | grep -v externalmodels)
GOTEST_DIRECTORIES       ?= $$(find ./internal/ -type f -iname "*_test.go" -exec dirname {} \; | uniq)

export GO111MODULE=on
export GOPRIVATE=github.com/form3tech-oss
export GOFLAGS=-mod=vendor

DOCKER_IMG ?= form3tech/form3-toxies

.PHONY: build
build:
	@echo "==> Building..."
	@go install ./...

.PHONY: goimports
goimports: install-goimports
	goimports -w $(GOFMT_FILES)

.PHONY: install-goimports
install-goimports:
	@type goimports >/dev/null 2>&1 || (cd /tmp && go get golang.org/x/tools/cmd/goimports && cd -)

.PHONY: vendor
vendor:
	@go mod tidy && go mod vendor && go mod verify

.PHONY: publish
publish:
	@echo "==> Building docker image..."
	docker build --build-arg APPNAME=form3-toxies -f build/package/form3-toxies/Dockerfile -t $(DOCKER_IMG):$(TRAVIS_TAG) .
	@echo "==> Logging in to the docker registry..."
	echo "$(DOCKER_PASSWORD)" | docker login -u "$(DOCKER_USERNAME)" --password-stdin
	@echo "==> Pushing built image..."
	docker push $(DOCKER_IMG):$(TRAVIS_TAG)
	docker tag $(DOCKER_IMG):$(TRAVIS_TAG) $(DOCKER_IMG):latest
	docker push $(DOCKER_IMG):latest