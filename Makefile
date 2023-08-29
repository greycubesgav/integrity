BIN_DIR=./bin
BIN=integrity
INSTALL_LOCATION=/usr/local/bin
LINUX_OS=linux
MAC_OS=darwin
BSD_OS=freebsd
ARCHINTEL=amd64
ARCHARM=arm64
VERSION := $(shell $(BIN_DIR)/$(BIN) --version)

.PHONY: build build-debug test install clean release bin-dir version

all:
	@echo "Default target"

build: bin-dir
	go build -o $(BIN_DIR)/$(BIN);

build-symlinks:
	ln -s ./integrity ./bin/integrity.sha1
	ln -s ./integrity ./bin/integrity.md5
	ln -s ./integrity ./bin/integrity.sha256
	ln -s ./integrity ./bin/integrity.sha512
	ln -s ./integrity ./bin/integrity.phash
	ln -s ./integrity ./bin/integrity.oshash

build-linux-intel: bin-dir
	GOOS=$(LINUX_OS) GOARCH=$(ARCHINTEL) go build -o $(BIN_DIR)/$(BIN)_$(LINUX_OS)_$(ARCHINTEL) -ldflags "-s -w -extldflags \"-static\""

build-linux-arm: bin-dir
	GOOS=$(LINUX_OS) GOARCH=$(ARCHARM) go build -o $(BIN_DIR)/$(BIN)_$(LINUX_OS)_$(ARCHARM) -ldflags "-extldflags \"-static\""

build-darwin-intel: bin-dir
	GOOS=$(MAC_OS) GOARCH=$(ARCHINTEL) go build -o $(BIN_DIR)/$(BIN)_$(MAC_OS)_$(ARCHINTEL);

build-darwin-arm: bin-dir
	GOOS=$(MAC_OS) GOARCH=$(ARCHARM) go build -o $(BIN_DIR)/$(BIN)_$(MAC_OS)_$(ARCHARM);

build-bsd-intel: bin-dir
	GOOS=$(BSD_OS) GOARCH=$(ARCHINTEL) go build -o $(BIN_DIR)/$(BIN)_$(BSD_OS)_$(ARCHINTEL);

build-bsd-arm: bin-dir
	GOOS=$(BSD_OS) GOARCH=$(ARCHARM) go build -o $(BIN_DIR)/$(BIN)_$(BSD_OS)_$(ARCHARM);

build-all: build-symlinks build-linux-intel build-linux-arm build-darwin-intel build-darwin-arm build-bsd-intel build-bsd-arm

package-deb-intel: package-deb-control
	fpm -s dir -t deb -n integrity -v $(VERSION) -a $(ARCHINTEL) --deb-custom-control ./packaging/debian/control -C ./bin/ integrity_linux_$(ARCHINTEL)=$(INSTALL_LOCATION)/$(BIN) \
	integrity.sha1=$(INSTALL_LOCATION)/$(BIN).sha1 \
	integrity.md5=$(INSTALL_LOCATION)/$(BIN).md5 \
	integrity.sha256=$(INSTALL_LOCATION)/$(BIN).sha256 \
	integrity.sha512=$(INSTALL_LOCATION)/$(BIN).sha512 \
	integrity.phash=$(INSTALL_LOCATION)/$(BIN).phash \
	integrity.oshash=$(INSTALL_LOCATION)/$(BIN).oshash

package-deb-arm: package-deb-control
	fpm -s dir -t deb -n integrity -v $(VERSION) -a $(ARCHARM) -C ./bin/ integrity_linux_$(ARCHARM)=$(INSTALL_LOCATION)/$(BIN) \
	integrity.sha1=$(INSTALL_LOCATION)/$(BIN).sha1 \
	integrity.md5=$(INSTALL_LOCATION)/$(BIN).md5 \
	integrity.sha256=$(INSTALL_LOCATION)/$(BIN).sha256 \
	integrity.sha512=$(INSTALL_LOCATION)/$(BIN).sha512 \
	integrity.phash=$(INSTALL_LOCATION)/$(BIN).phash \
	integrity.oshash=$(INSTALL_LOCATION)/$(BIN).oshash

package-rpm-intel:
	fpm -s dir -t rpm -n integrity -v $(VERSION) -a $(ARCHINTEL) -C ./bin/ integrity_linux_$(ARCHINTEL)=$(INSTALL_LOCATION)/$(BIN) \
	integrity.sha1=$(INSTALL_LOCATION)/$(BIN).sha1 \
	integrity.md5=$(INSTALL_LOCATION)/$(BIN).md5 \
	integrity.sha256=$(INSTALL_LOCATION)/$(BIN).sha256 \
	integrity.sha512=$(INSTALL_LOCATION)/$(BIN).sha512 \
	integrity.phash=$(INSTALL_LOCATION)/$(BIN).phash \
	integrity.oshash=$(INSTALL_LOCATION)/$(BIN).oshash

package-rpm-arm:
	fpm -s dir -t rpm -n integrity -v $(VERSION) -a $(ARCHARM) -C ./bin/ integrity_linux_$(ARCHARM)=$(INSTALL_LOCATION)/$(BIN) \
	integrity.sha1=$(INSTALL_LOCATION)/$(BIN).sha1 \
	integrity.md5=$(INSTALL_LOCATION)/$(BIN).md5 \
	integrity.sha256=$(INSTALL_LOCATION)/$(BIN).sha256 \
	integrity.sha512=$(INSTALL_LOCATION)/$(BIN).sha512 \
	integrity.phash=$(INSTALL_LOCATION)/$(BIN).phash \
	integrity.oshash=$(INSTALL_LOCATION)/$(BIN).oshash

package-apk-intel:
	fpm -s dir -t apk -n integrity -v $(VERSION) -a $(ARCHINTEL) -C ./bin/ integrity_linux_$(ARCHINTEL)=$(INSTALL_LOCATION)/$(BIN) \
	integrity.sha1=$(INSTALL_LOCATION)/$(BIN).sha1 \
	integrity.md5=$(INSTALL_LOCATION)/$(BIN).md5 \
	integrity.sha256=$(INSTALL_LOCATION)/$(BIN).sha256 \
	integrity.sha512=$(INSTALL_LOCATION)/$(BIN).sha512 \
	integrity.phash=$(INSTALL_LOCATION)/$(BIN).phash \
	integrity.oshash=$(INSTALL_LOCATION)/$(BIN).oshash

package-apk-arm:
	fpm -s dir -t apk -n integrity -v $(VERSION) -a $(ARCHARM) -C ./bin/ integrity_linux_$(ARCHARM)=$(INSTALL_LOCATION)/$(BIN) \
	integrity.sha1=$(INSTALL_LOCATION)/$(BIN).sha1 \
	integrity.md5=$(INSTALL_LOCATION)/$(BIN).md5 \
	integrity.sha256=$(INSTALL_LOCATION)/$(BIN).sha256 \
	integrity.sha512=$(INSTALL_LOCATION)/$(BIN).sha512 \
	integrity.phash=$(INSTALL_LOCATION)/$(BIN).phash \
	integrity.oshash=$(INSTALL_LOCATION)/$(BIN).oshash

package-slackware-intel:
	cd packaging/slackware
	./integrity.SlackBuild

package-freebsd:
	fpm -s dir -t apk -n freebsd -v 1.0.0 -a $(ARCHINTEL) ./

package-all: package-deb-intel package-deb-arm package-rpm-intel package-rpm-arm package-apk-intel package-apk-arm

package-deb-control:
	sed -i 's/<version>/$(VERSION)/g' ./packaging/debian/control
	sed -i 's/<arch>/$(ARCHINTEL)/g' ./packaging/debian/control

bin-dir:
	mkdir -p $(BIN_DIR)


