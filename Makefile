REGISTRY ?= ghcr.io
OWNER ?= racklet
IMAGE_NAME ?= render-drawio-action
IMAGE := $(REGISTRY)/$(OWNER)/$(IMAGE_NAME)
TAG ?=
# Test using a file with a whitespace in it
TEST_FILE_BASE ?= "test/foo 1"

GIT_VERSION := $(shell git describe --tags HEAD 2>/dev/null || echo "v0.0.0")
GIT_STATUS := $(if $(shell git status -s),-dirty)
GIT_TAG := $(GIT_VERSION)$(GIT_STATUS)
GIT_MAJOR := $(shell echo "$(GIT_VERSION)" | cut -d. -f1)

all: build

build:
	docker build -t $(IMAGE):$(GIT_TAG) --pull .

push: build
	docker push $(IMAGE):$(GIT_TAG)
ifneq ($(TAG),)
	docker tag $(IMAGE):$(GIT_TAG) $(IMAGE):$(TAG)
	docker push $(IMAGE):$(TAG)
endif

push-major: push
ifeq ($(GIT_STATUS),)
	docker tag "$(IMAGE):$(GIT_TAG)" "$(IMAGE):$(GIT_MAJOR)"
	docker push "$(IMAGE):$(GIT_MAJOR)"
endif

test: build
	docker run -it -v $(shell pwd):/files $(IMAGE):$(GIT_TAG) --formats=png,svg,pdf,jpg
	$(MAKE) verify

verify:
	[ -f $(TEST_FILE_BASE).svg ] && [ -f $(TEST_FILE_BASE).png ] && \
	[ -f $(TEST_FILE_BASE).pdf ] && [ -f $(TEST_FILE_BASE).jpg ] && \
	echo "Files generated!" || (echo "e2e files not generated" && false)
