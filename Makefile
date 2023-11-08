BUILDDIR ?= $(CURDIR)/build
DOCKER := $(shell which docker)

export GO111MODULE = on

###############################################################################
###                               Build flags                               ###
###############################################################################

build_tags = netgo

# These lines here are essential to include the muslc library for static linking of libraries
# (which is needed for the wasmvm one) available during the build. Without them, the build will fail.
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

ldflags =
ifeq ($(LINK_STATICALLY),true)
  ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif

ldflags := $(strip $(ldflags))
BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

###############################################################################
###                                 Build                                   ###
###############################################################################

BUILD_TARGETS := build

build: BUILD_ARGS=-o $(BUILDDIR)/

create-builder: go.sum
	$(MAKE) -C contrib/images builder CONTEXT=$(CURDIR)

build-alpine: create-builder
	mkdir -p $(BUILDDIR)
	$(DOCKER) build -f Dockerfile --rm --tag kilem/desmos-loading .
	$(DOCKER) create --name desmos-loading --rm kilem/desmos-loading
	$(DOCKER) cp desmos-loading:/usr/bin/desmos-loading $(BUILDDIR)/desmos-loading
	$(DOCKER) rm desmos-loading

$(BUILD_TARGETS): go.sum $(BUILDDIR)/
	go $@ -mod=readonly $(BUILD_FLAGS) $(BUILD_ARGS) ./...

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

.PHONY: build