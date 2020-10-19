# These are generally passed to us by xcode, but we set working defaults for standalone compilation too.
ARCHS ?= arm64 armv7
CONFIGURATION_BUILD_DIR ?= $(CURDIR)/out
CONFIGURATION_TEMP_DIR ?= $(CURDIR)/.tmp

export CC ?= clang
LIPO ?= lipo
DESTDIR ?= $(CONFIGURATION_BUILD_DIR)
BUILDDIR ?= $(CONFIGURATION_TEMP_DIR)/libhcash

CFLAGS_PREFIX := $(if $(DEPLOYMENT_TARGET_CLANG_FLAG_NAME),-$(DEPLOYMENT_TARGET_CLANG_FLAG_NAME)=$($(DEPLOYMENT_TARGET_CLANG_ENV_NAME)),) -Wno-unused-command-line-argument -arch
GOARCH_arm64 := arm64
GOARCH_armv7 := arm
GOARCH_x86_64 := amd64

build: $(DESTDIR)/libhcash.a

REAL_GOROOT := $(shell go env GOROOT 2>/dev/null)
export GOROOT := $(BUILDDIR)/goroot
$(GOROOT)/.prepared:
	[ -n "$(REAL_GOROOT)" ]
	mkdir -p "$(GOROOT)"
	rsync -a --delete --exclude=pkg/obj/go-build "$(REAL_GOROOT)/" "$(GOROOT)/"
	touch "$@"

define libhcash-a
$(BUILDDIR)/libhcash-$(1).a: export CGO_ENABLED := 1
$(BUILDDIR)/libhcash-$(1).a: export CGO_CFLAGS := $(CFLAGS_PREFIX) $(ARCH)
$(BUILDDIR)/libhcash-$(1).a: export CGO_LDFLAGS := $(CFLAGS_PREFIX) $(ARCH)
$(BUILDDIR)/libhcash-$(1).a: export GOOS := darwin
$(BUILDDIR)/libhcash-$(1).a: export GOARCH := $(GOARCH_$(1))
$(BUILDDIR)/libhcash-$(1).a: $(GOROOT)/.prepared go.mod
	go build -tags ios -ldflags=-w -trimpath -v -o "$(BUILDDIR)/libhcash-$(1).a" -buildmode c-archive
	rm -f "$(BUILDDIR)/libhcash-$(1).h"
endef
$(foreach ARCH,$(ARCHS),$(eval $(call libhcash-a,$(ARCH))))


$(DESTDIR)/libhcash.a: $(foreach ARCH,$(ARCHS),$(BUILDDIR)/libhcash-$(ARCH).a)
	@mkdir -vp "$(DESTDIR)"
	$(LIPO) -create -output "$@" $^

clean:
	rm -rf "$(BUILDDIR)" "$(DESTDIR)/libhcash.a"

install: build

.PHONY: clean build install
