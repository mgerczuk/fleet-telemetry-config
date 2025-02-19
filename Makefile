APP_NAME := fleet-telemetry-config
ARCHS := amd64 arm

GITVERSION := $(shell git describe --tags --long)
VERSION := $(shell echo $(GITVERSION) | sed -E 's/v([0-9]+\.[0-9]+)-([0-9]+)-g.*/\1.\2/')

BUILD_DIR := build
DEB_DIR := $(BUILD_DIR)/deb

# Standard GO Einstellungen
export GO111MODULE = on

# Cross-Compilation Flags
LDFLAGS := -s -w

all: clean build package

build:
	@mkdir -p $(BUILD_DIR)
	@for arch in $(ARCHS); do \
		echo "Building for $$arch..."; \
		GOOS=linux GOARCH=$$arch go build -o $(BUILD_DIR)/$(APP_NAME)_$$arch -ldflags "$(LDFLAGS) -X main.version=$(GITVERSION)" cmd/main.go ; \
	done

package:
	@for arch in $(ARCHS); do \
		if [ "$$arch" = "arm" ]; then \
			deb_arch="armhf"; \
		else \
			deb_arch="$$arch"; \
		fi; \
		echo "Creating DEB package for $$arch (DEB arch: $$deb_arch)..."; \
		mkdir -p $(DEB_DIR)/DEBIAN; \
		mkdir -p $(DEB_DIR)/usr/bin; \
		cp $(BUILD_DIR)/$(APP_NAME)_$$arch $(DEB_DIR)/usr/bin/$(APP_NAME); \
		chmod 755 $(DEB_DIR)/usr/bin/$(APP_NAME); \
		echo "Package: $(APP_NAME)" > $(DEB_DIR)/DEBIAN/control; \
		echo "Version: $(VERSION)" >> $(DEB_DIR)/DEBIAN/control; \
		echo "Architecture: $$deb_arch" >> $(DEB_DIR)/DEBIAN/control; \
		echo "Maintainer: Martin Gerczuk <martin@mgsoftware.de>" >> $(DEB_DIR)/DEBIAN/control; \
		echo "Description: Tesla Fleet Telemetry Configuration" >> $(DEB_DIR)/DEBIAN/control; \
		dpkg-deb --build $(DEB_DIR) $(BUILD_DIR)/$(APP_NAME)_$(VERSION)_$$arch.deb; \
		rm -rf $(DEB_DIR); \
	done

clean:
	@rm -rf $(BUILD_DIR)
