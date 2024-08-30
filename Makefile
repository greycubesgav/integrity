BIN_DIR=./bin
PKGS_DIR=./pkgs
BIN=integrity
INSTALL_LOCATION=/usr/local/bin
MAINTAINER=Gavin Brown
EMAIL=integrity@greycubes.net
DESCRIPTION=integrity (CLI tool to calculate, store and verify a file's checksum stored with the file's extended attributes)
URL=https://github.com/greycubesgav/integrity/
LICENSE=LGPL
LINUX_OS=linux
MAC_OS=darwin
BSD_OS=freebsd
ARCHINTEL=amd64
ARCHARM=arm64
VERSION := $(shell grep integrity_version pkg/integrity/version.go | sed -n 's/.*"\([^"]*\)".*/\1/p')
PWD := $(shell pwd)

.PHONY: build build-debug test install clean release version docker-build-image docker-build-image

all: build
	@echo "Default target"

go-get-all:
	go get -d ./...

build: bin
	go build -o bin/integrity cmd/integrity/integrity.go;

bin:
	mkdir -p $(BIN_DIR)

build-symlinks:
	ln -sf ./integrity ./bin/integrity.sha1
	ln -sf ./integrity ./bin/integrity.md5
	ln -sf ./integrity ./bin/integrity.sha256
	ln -sf ./integrity ./bin/integrity.sha512
	ln -sf ./integrity ./bin/integrity.phash
	ln -sf ./integrity ./bin/integrity.oshash

build-linux-intel: bin build-symlinks
	@if [ -f "$(BIN_DIR)/$(BIN)_$(LINUX_OS)_$(ARCHINTEL)" ]; then \
			echo "File exists, skipping step."; \
	else \
			echo "File does not exist, running step."; \
			GOOS=$(LINUX_OS) GOARCH=$(ARCHINTEL) go build -o $(BIN_DIR)/$(BIN)_$(LINUX_OS)_$(ARCHINTEL) -ldflags "-s -w -extldflags \"-static\"" cmd/integrity/integrity.go; \
	fi

build-linux-arm: bin build-symlinks
	GOOS=$(LINUX_OS) GOARCH=$(ARCHARM) go build -o $(BIN_DIR)/$(BIN)_$(LINUX_OS)_$(ARCHARM) -ldflags "-extldflags \"-static\"" cmd/integrity/integrity.go

build-darwin-intel: bin build-symlinks
	GOOS=$(MAC_OS) GOARCH=$(ARCHINTEL) go build -o $(BIN_DIR)/$(BIN)_$(MAC_OS)_$(ARCHINTEL) cmd/integrity/integrity.go;

build-darwin-arm: bin build-symlinks
	GOOS=$(MAC_OS) GOARCH=$(ARCHARM) go build -o $(BIN_DIR)/$(BIN)_$(MAC_OS)_$(ARCHARM) cmd/integrity/integrity.go;

build-bsd-intel: bin build-symlinks
	GOOS=$(BSD_OS) GOARCH=$(ARCHINTEL) go build -o $(BIN_DIR)/$(BIN)_$(BSD_OS)_$(ARCHINTEL) cmd/integrity/integrity.go;

build-bsd-arm: bin build-symlinks
	GOOS=$(BSD_OS) GOARCH=$(ARCHARM) go build -o $(BIN_DIR)/$(BIN)_$(BSD_OS)_$(ARCHARM) cmd/integrity/integrity.go;

build-all: bin build-symlinks build-linux-intel build-linux-arm build-darwin-intel build-darwin-arm build-bsd-intel build-bsd-arm

pkgs:
	mkdir -p $(PKGS_DIR)

package-all: package-deb-intel package-deb-arm package-rpm-intel package-rpm-arm package-apk-intel package-apk-arm package-tar-intel package-tar-arm package-slackware-intel package-slackware-arm

package-deb-intel: build-linux-intel pkgs package-deb-control package-deb-control-intel
	fpm -s dir -t deb -n integrity -p pkgs -v $(VERSION) -a $(ARCHINTEL) --deb-custom-control ./packaging/debian/control  --license "$(LICENSE)" --description="$(DESCRIPTION)" -m "$(MAINTAINER)" --url "$(URL)" -C ./bin/  integrity_linux_$(ARCHINTEL)=$(INSTALL_LOCATION)/$(BIN) \
	integrity.sha1=$(INSTALL_LOCATION)/$(BIN).sha1 \
	integrity.md5=$(INSTALL_LOCATION)/$(BIN).md5 \
	integrity.sha256=$(INSTALL_LOCATION)/$(BIN).sha256 \
	integrity.sha512=$(INSTALL_LOCATION)/$(BIN).sha512 \
	integrity.phash=$(INSTALL_LOCATION)/$(BIN).phash \
	integrity.oshash=$(INSTALL_LOCATION)/$(BIN).oshash

package-deb-arm: build-linux-arm pkgs package-deb-control package-deb-control-arm
	fpm -s dir -t deb -n integrity -p pkgs -v $(VERSION) -a $(ARCHARM) --deb-custom-control ./packaging/debian/control  --license "$(LICENSE)" --description="$(DESCRIPTION)" -m "$(MAINTAINER)" --url "$(URL)" -C ./bin/  integrity_linux_$(ARCHARM)=$(INSTALL_LOCATION)/$(BIN) \
	integrity.sha1=$(INSTALL_LOCATION)/$(BIN).sha1 \
	integrity.md5=$(INSTALL_LOCATION)/$(BIN).md5 \
	integrity.sha256=$(INSTALL_LOCATION)/$(BIN).sha256 \
	integrity.sha512=$(INSTALL_LOCATION)/$(BIN).sha512 \
	integrity.phash=$(INSTALL_LOCATION)/$(BIN).phash \
	integrity.oshash=$(INSTALL_LOCATION)/$(BIN).oshash

package-rpm-intel: build-linux-intel pkgs
	fpm -s dir -t rpm -n integrity -p pkgs -v $(VERSION) -a $(ARCHINTEL) --license "$(LICENSE)" --description="$(DESCRIPTION)" -m "$(MAINTAINER)" --url "$(URL)" -C ./bin/ integrity_linux_$(ARCHINTEL)=$(INSTALL_LOCATION)/$(BIN) \
	integrity.sha1=$(INSTALL_LOCATION)/$(BIN).sha1 \
	integrity.md5=$(INSTALL_LOCATION)/$(BIN).md5 \
	integrity.sha256=$(INSTALL_LOCATION)/$(BIN).sha256 \
	integrity.sha512=$(INSTALL_LOCATION)/$(BIN).sha512 \
	integrity.phash=$(INSTALL_LOCATION)/$(BIN).phash \
	integrity.oshash=$(INSTALL_LOCATION)/$(BIN).oshash

package-rpm-arm: build-linux-arm pkgs
	fpm -s dir -t rpm -n integrity -p pkgs -v $(VERSION) -a $(ARCHARM)  --license "$(LICENSE)" --description="$(DESCRIPTION)" -m "$(MAINTAINER)" --url "$(URL)" -C ./bin/  integrity_linux_$(ARCHARM)=$(INSTALL_LOCATION)/$(BIN) \
	integrity.sha1=$(INSTALL_LOCATION)/$(BIN).sha1 \
	integrity.md5=$(INSTALL_LOCATION)/$(BIN).md5 \
	integrity.sha256=$(INSTALL_LOCATION)/$(BIN).sha256 \
	integrity.sha512=$(INSTALL_LOCATION)/$(BIN).sha512 \
	integrity.phash=$(INSTALL_LOCATION)/$(BIN).phash \
	integrity.oshash=$(INSTALL_LOCATION)/$(BIN).oshash

package-apk-intel: build-linux-intel pkgs
	fpm -s dir -t apk -n integrity -p pkgs -v $(VERSION) -a $(ARCHINTEL)  --license "$(LICENSE)" --description="$(DESCRIPTION)" -m "$(MAINTAINER)" --url "$(URL)" -C ./bin/  integrity_linux_$(ARCHINTEL)=$(INSTALL_LOCATION)/$(BIN) \
	integrity.sha1=$(INSTALL_LOCATION)/$(BIN).sha1 \
	integrity.md5=$(INSTALL_LOCATION)/$(BIN).md5 \
	integrity.sha256=$(INSTALL_LOCATION)/$(BIN).sha256 \
	integrity.sha512=$(INSTALL_LOCATION)/$(BIN).sha512 \
	integrity.phash=$(INSTALL_LOCATION)/$(BIN).phash \
	integrity.oshash=$(INSTALL_LOCATION)/$(BIN).oshash

package-apk-arm: build-linux-arm pkgs
	fpm -s dir -t apk -n integrity -p pkgs -v $(VERSION) -a $(ARCHARM)  --license "$(LICENSE)" --description="$(DESCRIPTION)" -m "$(MAINTAINER)" --url "$(URL)" -C ./bin/  integrity_linux_$(ARCHARM)=$(INSTALL_LOCATION)/$(BIN) \
	integrity.sha1=$(INSTALL_LOCATION)/$(BIN).sha1 \
	integrity.md5=$(INSTALL_LOCATION)/$(BIN).md5 \
	integrity.sha256=$(INSTALL_LOCATION)/$(BIN).sha256 \
	integrity.sha512=$(INSTALL_LOCATION)/$(BIN).sha512 \
	integrity.phash=$(INSTALL_LOCATION)/$(BIN).phash \
	integrity.oshash=$(INSTALL_LOCATION)/$(BIN).oshash

package-tar-intel: build-linux-intel pkgs
	fpm -s dir -t tar -n integrity_$(VERSION)_$(ARCHINTEL) -p pkgs -v $(VERSION) -a $(ARCHINTEL)  --license "$(LICENSE)" --description="$(DESCRIPTION)" -m "$(MAINTAINER)" --url "$(URL)" -C ./bin/  integrity_linux_$(ARCHINTEL)=$(INSTALL_LOCATION)/$(BIN) \
	integrity.sha1=$(INSTALL_LOCATION)/$(BIN).sha1 \
	integrity.md5=$(INSTALL_LOCATION)/$(BIN).md5 \
	integrity.sha256=$(INSTALL_LOCATION)/$(BIN).sha256 \
	integrity.sha512=$(INSTALL_LOCATION)/$(BIN).sha512 \
	integrity.phash=$(INSTALL_LOCATION)/$(BIN).phash \
	integrity.oshash=$(INSTALL_LOCATION)/$(BIN).oshash

package-tar-arm: build-linux-arm pkgs
	fpm -s dir -t tar -n integrity_$(VERSION)_$(ARCHARM) -p pkgs -v $(VERSION) -a $(ARCHARM)  --license "$(LICENSE)" --description="$(DESCRIPTION)" -m "$(MAINTAINER)" --url "$(URL)" -C ./bin/  integrity_linux_$(ARCHARM)=$(INSTALL_LOCATION)/$(BIN) \
	integrity.sha1=$(INSTALL_LOCATION)/$(BIN).sha1 \
	integrity.md5=$(INSTALL_LOCATION)/$(BIN).md5 \
	integrity.sha256=$(INSTALL_LOCATION)/$(BIN).sha256 \
	integrity.sha512=$(INSTALL_LOCATION)/$(BIN).sha512 \
	integrity.phash=$(INSTALL_LOCATION)/$(BIN).phash \
	integrity.oshash=$(INSTALL_LOCATION)/$(BIN).oshash

package-slackware-intel: build-linux-intel pkgs package-slackware-info
  export PATH="$(pwd)/packaging/slackware/:$PATH" && \
	cd ./packaging/slackware && \
	ARCH='$(ARCHINTEL)' VERSION='$(VERSION)' OUTPUT="$$(pwd)/../../pkgs/" ./integrity.SlackBuild

package-slackware-arm: build-linux-intel pkgs package-slackware-info
	cd ./packaging/slackware && \
	cp makepkg /sbin/makepkg && \
	ARCH='$(ARCHARM)' VERSION='$(VERSION)' OUTPUT="$$(pwd)/../../pkgs/" ./integrity.SlackBuild

docker-package-slackware-intel: build-linux-intel pkgs docker-build-slackware-image
	docker cp $(docker create --name tc "greycubesgav/integrity-slackware-build:$(VERSION)"):/tmp/integrity-$(VERSION)-x86_64-1_GG.tgz ./pkgs && docker rm tc

package-freebsd: build-bsd-intel pkgs
	fpm -s dir -t apk -n freebsd -v $(VERSION) -a $(ARCHINTEL) ./

package-slackware-info:
	sed -i 's|<version>|$(VERSION)|g' ./packaging/slackware/integrity.info
	sed -i 's|<homepage>|$(URL)|g' ./packaging/slackware/integrity.info
	sed -i 's|<maintainer>|$(MAINTAINER)|g' ./packaging/slackware/integrity.info
	sed -i 's|<email>|$(EMAIL)|g' ./packaging/slackware/integrity.info

package-deb-control:
	sed -i 's|<version>|$(VERSION)|g' ./packaging/debian/control
	sed -i 's|<homepage>|$(URL)|g' ./packaging/debian/control
	sed -i 's|<maintainer>|$(MAINTAINER)|g' ./packaging/debian/control

package-deb-control-intel:
	sed -i 's/<arch>/$(ARCHINTEL)/g' ./packaging/debian/control

package-deb-control-arm:
	sed -i 's/<arch>/$(ARCHARM)/g' ./packaging/debian/control

docker-build-image:
	docker build -t "greycubesgav/integrity-build:$(VERSION)" -f Dockerfile .

docker-dev-image:
	docker run -it --rm -v "$(PWD)":/usr/src/myapp -w /usr/src/myapp greycubesgav/integrity-build:$(VERSION)

docker-build-slackware-image:
	docker build -t "greycubesgav/integrity-slackware-build:$(VERSION)" -f Dockerfile.slackware .

docker-dev-slackware-image:
	docker run -it --rm -v "$(PWD)":/usr/src/myapp -w /usr/src/myapp greycubesgav/integrity-slackware-build:$(VERSION)

build-in-docker:
	docker run --rm -v "$(PWD)":/usr/src/myapp -w /usr/src/myapp golang:1.16 bash -c "go get -d ./...  ; go build -v"

test:
	#go test github.com/greycubesgav/integrity/pkg/integrity
	go test ./pkg/integrity

test-code-coverage:
	# Run to generate code coverage, then cmd-shift-p : go:toggle test coverage to view code coverage
	go test -v -cover ./...

test-add-linux-attr:
	setfattr -n user.test -v "This is the user.test attribute" data.dat
	setfattr -n test -v "This is the test test attribute" data.dat

test-list-linux-attr:
	getfattr -d data.dat

test-make-data.dat:
	echo 'hello world' > data.dat

test-github-package:
	act push -j package

show-version:
	@echo $(VERSION)
