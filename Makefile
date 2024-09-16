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

clean-bin:
	rm -rf $(BIN_DIR);

clean-pkgs:
	rm -rf $(PKGS_DIR);

clean-all: clean-pkgs clean-bin

go-get-all:
	go get -d ./...

build: bin
	go build -o bin/integrity cmd/integrity/integrity.go;

bin:
	@if [ ! -d $(BIN_DIR) ]; then \
	  echo "Creating bin directory at $(BIN_DIR)"; \
		mkdir -p $(BIN_DIR); \
	fi

build-symlinks: bin
	@for checksum in md5 sha1 sha256 sha512 phash oshash; do \
		link=$(BIN_DIR)/$(BIN).$$checksum; \
		if [ ! -L "$$link" ]; then \
			echo "Creating symlink for $$link"; \
			ln -sf ./$(BIN) $$link; \
		fi; \
	done ;

$(BIN_DIR)/$(BIN)_$(LINUX_OS)_$(ARCHINTEL):
	@echo "Creating $(BIN_DIR)/$(BIN)_$(LINUX_OS)_$(ARCHINTEL) ..."
	GOOS=$(LINUX_OS) GOARCH=$(ARCHINTEL) go build -o $(BIN_DIR)/$(BIN)_$(LINUX_OS)_$(ARCHINTEL) -ldflags "-extldflags \"-static\"" cmd/integrity/integrity.go

$(BIN_DIR)/$(BIN)_$(LINUX_OS)_$(ARCHARM):
	@echo "Creating $(BIN_DIR)/$(BIN)_$(LINUX_OS)_$(ARCHARM) ..."
	GOOS=$(LINUX_OS) GOARCH=$(ARCHARM) go build -o $(BIN_DIR)/$(BIN)_$(LINUX_OS)_$(ARCHARM) -ldflags "-extldflags \"-static\"" cmd/integrity/integrity.go

$(BIN_DIR)/$(BIN)_$(MAC_OS)_$(ARCHINTEL):
	@echo "Creating $(BIN_DIR)/$(BIN)_$(MAC_OS)_$(ARCHINTEL) ..."
	GOOS=$(MAC_OS) GOARCH=$(ARCHINTEL) go build -o $(BIN_DIR)/$(BIN)_$(MAC_OS)_$(ARCHINTEL) cmd/integrity/integrity.go;

$(BIN_DIR)/$(BIN)_$(MAC_OS)_$(ARCHARM):
	@echo "Creating $(BIN_DIR)/$(BIN)_$(MAC_OS)_$(ARCHARM) ..."
	GOOS=$(MAC_OS) GOARCH=$(ARCHARM) go build -o $(BIN_DIR)/$(BIN)_$(MAC_OS)_$(ARCHARM) cmd/integrity/integrity.go;

$(BIN_DIR)/$(BIN)_$(BSD_OS)_$(ARCHINTEL):
	@echo "Creating $(BIN_DIR)/$(BIN)_$(BSD_OS)_$(ARCHINTEL) ..."
	GOOS=$(BSD_OS) GOARCH=$(ARCHINTEL) go build -o $(BIN_DIR)/$(BIN)_$(BSD_OS)_$(ARCHINTEL) cmd/integrity/integrity.go;

$(BIN_DIR)/$(BIN)_$(BSD_OS)_$(ARCHARM):
	@echo "Creating $(BIN_DIR)/$(BIN)_$(BSD_OS)_$(ARCHARM) ..."
	GOOS=$(BSD_OS) GOARCH=$(ARCHARM) go build -o $(BIN_DIR)/$(BIN)_$(BSD_OS)_$(ARCHARM) cmd/integrity/integrity.go;

build-all: $(BIN_DIR)/$(BIN)_$(LINUX_OS)_$(ARCHINTEL) $(BIN_DIR)/$(BIN)_$(LINUX_OS)_$(ARCHARM) $(BIN_DIR)/$(BIN)_$(MAC_OS)_$(ARCHINTEL) $(BIN_DIR)/$(BIN)_$(MAC_OS)_$(ARCHARM) $(BIN_DIR)/$(BIN)_$(BSD_OS)_$(ARCHINTEL) $(BIN_DIR)/$(BIN)_$(BSD_OS)_$(ARCHARM)

pkgs:
	mkdir -p $(PKGS_DIR)

package-all: build-all build-symlinks package-deb-intel package-deb-arm package-rpm-intel package-rpm-arm package-apk-intel package-apk-arm package-tar-intel package-tar-arm package-slackware-intel package-slackware-arm

package-deb-intel: $(BIN_DIR)/$(BIN)_$(LINUX_OS)_$(ARCHINTEL) pkgs
	@if [ -f $(PKGS_DIR)/$(BIN)_$(VERSION)_$(ARCHINTEL).deb ]; then \
		echo "$(PKGS_DIR)/$(BIN)_$(VERSION)_$(ARCHINTEL).deb already exists"; \
	else \
		$(MAKE) build-symlinks && \
		$(MAKE) package-deb-control && \
		$(MAKE) package-deb-control-intel && \
		fpm -s dir -t deb -n $(BIN) -p pkgs -v $(VERSION) -a $(ARCHINTEL) --deb-custom-control ./packaging/debian/control  --license "$(LICENSE)" --description="$(DESCRIPTION)" -m "$(MAINTAINER)" --url "$(URL)" -C ./bin/  $(BIN)_$(LINUX_OS)_$(ARCHINTEL)=$(INSTALL_LOCATION)/$(BIN) \
		integrity.sha1=$(INSTALL_LOCATION)/$(BIN).sha1 \
		integrity.md5=$(INSTALL_LOCATION)/$(BIN).md5 \
		integrity.sha256=$(INSTALL_LOCATION)/$(BIN).sha256 \
		integrity.sha512=$(INSTALL_LOCATION)/$(BIN).sha512 \
		integrity.phash=$(INSTALL_LOCATION)/$(BIN).phash \
		integrity.oshash=$(INSTALL_LOCATION)/$(BIN).oshash ;\
	fi

package-deb-arm: $(BIN_DIR)/$(BIN)_$(LINUX_OS)_$(ARCHARM) pkgs
	@if [ -f $(PKGS_DIR)/$(BIN)_$(VERSION)_$(ARCHARM).deb ]; then \
		echo "$(PKGS_DIR)/$(BIN)_$(VERSION)_$(ARCHARM).deb already exists"; \
	else \
		$(MAKE) build-symlinks && \
		$(MAKE) package-deb-control && \
		$(MAKE) package-deb-control-arm && \
		fpm -s dir -t deb -n $(BIN) -p pkgs -v $(VERSION) -a $(ARCHARM) \
		--deb-custom-control ./packaging/debian/control --license "$(LICENSE)" \
		--description="$(DESCRIPTION)" -m "$(MAINTAINER)" --url "$(URL)" \
		-C $(BIN_DIR)/ $(BIN)_$(LINUX_OS)_$(ARCHARM)=$(INSTALL_LOCATION)/$(BIN) \
		$(BIN).sha1=$(INSTALL_LOCATION)/$(BIN).sha1 \
		$(BIN).md5=$(INSTALL_LOCATION)/$(BIN).md5 \
		$(BIN).sha256=$(INSTALL_LOCATION)/$(BIN).sha256 \
		$(BIN).sha512=$(INSTALL_LOCATION)/$(BIN).sha512 \
		$(BIN).phash=$(INSTALL_LOCATION)/$(BIN).phash \
		$(BIN).oshash=$(INSTALL_LOCATION)/$(BIN).oshash; \
	fi

package-rpm-intel: $(BIN_DIR)/$(BIN)_$(LINUX_OS)_$(ARCHINTEL) pkgs
	@if [ -f $(PKGS_DIR)/$(BIN)-$(VERSION)-1.x86_64.rpm ]; then \
		echo "$(PKGS_DIR)/$(BIN)-$(VERSION)-1.x86_64.rpm already exists"; \
	else \
		$(MAKE) build-symlinks && \
		fpm -s dir -t rpm -n $(BIN) -p pkgs -v $(VERSION) -a $(ARCHINTEL) \
		--license "$(LICENSE)" --description="$(DESCRIPTION)" -m "$(MAINTAINER)" --url "$(URL)" \
		-C $(BIN_DIR)/ $(BIN)_$(LINUX_OS)_$(ARCHINTEL)=$(INSTALL_LOCATION)/$(BIN) \
		$(BIN).sha1=$(INSTALL_LOCATION)/$(BIN).sha1 \
		$(BIN).md5=$(INSTALL_LOCATION)/$(BIN).md5 \
		$(BIN).sha256=$(INSTALL_LOCATION)/$(BIN).sha256 \
		$(BIN).sha512=$(INSTALL_LOCATION)/$(BIN).sha512 \
		$(BIN).phash=$(INSTALL_LOCATION)/$(BIN).phash \
		$(BIN).oshash=$(INSTALL_LOCATION)/$(BIN).oshash; \
	fi

package-rpm-arm: $(BIN_DIR)/$(BIN)_$(LINUX_OS)_$(ARCHARM) pkgs
	@if [ -f $(PKGS_DIR)/$(BIN)-$(VERSION)-1.aarch64.rpm ]; then \
		echo "$(PKGS_DIR)/$(BIN)-$(VERSION)-1.aarch64.rpm already exists"; \
	else \
		$(MAKE) build-symlinks && \
		fpm -s dir -t rpm -n $(BIN) -p pkgs -v $(VERSION) -a $(ARCHARM) \
		--license "$(LICENSE)" --description="$(DESCRIPTION)" -m "$(MAINTAINER)" --url "$(URL)" \
		-C $(BIN_DIR)/ $(BIN)_$(LINUX_OS)_$(ARCHARM)=$(INSTALL_LOCATION)/$(BIN) \
		$(BIN).sha1=$(INSTALL_LOCATION)/$(BIN).sha1 \
		$(BIN).md5=$(INSTALL_LOCATION)/$(BIN).md5 \
		$(BIN).sha256=$(INSTALL_LOCATION)/$(BIN).sha256 \
		$(BIN).sha512=$(INSTALL_LOCATION)/$(BIN).sha512 \
		$(BIN).phash=$(INSTALL_LOCATION)/$(BIN).phash \
		$(BIN).oshash=$(INSTALL_LOCATION)/$(BIN).oshash; \
	fi


package-apk-intel: $(BIN_DIR)/$(BIN)_$(LINUX_OS)_$(ARCHINTEL) pkgs
	@if [ -f $(PKGS_DIR)/$(BIN)_$(VERSION)_$(ARCHINTEL).apk ]; then \
		echo "$(PKGS_DIR)/$(BIN)_$(VERSION)_$(ARCHINTEL).apk already exists"; \
	else \
		$(MAKE) build-symlinks && \
		fpm -s dir -t apk -n $(BIN) -p pkgs -v $(VERSION) -a $(ARCHINTEL) \
		--license "$(LICENSE)" --description="$(DESCRIPTION)" -m "$(MAINTAINER)" --url "$(URL)" \
		-C $(BIN_DIR)/ $(BIN)_$(LINUX_OS)_$(ARCHINTEL)=$(INSTALL_LOCATION)/$(BIN) \
		$(BIN).sha1=$(INSTALL_LOCATION)/$(BIN).sha1 \
		$(BIN).md5=$(INSTALL_LOCATION)/$(BIN).md5 \
		$(BIN).sha256=$(INSTALL_LOCATION)/$(BIN).sha256 \
		$(BIN).sha512=$(INSTALL_LOCATION)/$(BIN).sha512 \
		$(BIN).phash=$(INSTALL_LOCATION)/$(BIN).phash \
		$(BIN).oshash=$(INSTALL_LOCATION)/$(BIN).oshash; \
	fi

package-apk-arm: $(BIN_DIR)/$(BIN)_$(LINUX_OS)_$(ARCHARM) pkgs
	@if [ -f $(PKGS_DIR)/$(BIN)_$(VERSION)_$(ARCHARM).apk ]; then \
		echo "$(PKGS_DIR)/$(BIN)_$(VERSION)_$(ARCHARM).apk already exists"; \
	else \
		$(MAKE) build-symlinks && \
		fpm -s dir -t apk -n $(BIN) -p pkgs -v $(VERSION) -a $(ARCHARM) \
		--license "$(LICENSE)" --description="$(DESCRIPTION)" -m "$(MAINTAINER)" --url "$(URL)" \
		-C $(BIN_DIR)/ $(BIN)_$(LINUX_OS)_$(ARCHARM)=$(INSTALL_LOCATION)/$(BIN) \
		$(BIN).sha1=$(INSTALL_LOCATION)/$(BIN).sha1 \
		$(BIN).md5=$(INSTALL_LOCATION)/$(BIN).md5 \
		$(BIN).sha256=$(INSTALL_LOCATION)/$(BIN).sha256 \
		$(BIN).sha512=$(INSTALL_LOCATION)/$(BIN).sha512 \
		$(BIN).phash=$(INSTALL_LOCATION)/$(BIN).phash \
		$(BIN).oshash=$(INSTALL_LOCATION)/$(BIN).oshash; \
	fi

package-tar-intel: $(BIN_DIR)/$(BIN)_$(LINUX_OS)_$(ARCHINTEL) pkgs
	@if [ -f $(PKGS_DIR)/$(BIN)_$(VERSION)_$(ARCHINTEL).tar ]; then \
		echo "$(PKGS_DIR)/$(BIN)_$(VERSION)_$(ARCHINTEL).tar already exists"; \
	else \
		$(MAKE) build-symlinks && \
		fpm -s dir -t tar -n $(BIN)_$(VERSION)_$(ARCHINTEL) -p pkgs -v $(VERSION) -a $(ARCHINTEL) \
		--license "$(LICENSE)" --description="$(DESCRIPTION)" -m "$(MAINTAINER)" --url "$(URL)" \
		-C $(BIN_DIR)/ $(BIN)_$(LINUX_OS)_$(ARCHINTEL)=$(INSTALL_LOCATION)/$(BIN) \
		$(BIN).sha1=$(INSTALL_LOCATION)/$(BIN).sha1 \
		$(BIN).md5=$(INSTALL_LOCATION)/$(BIN).md5 \
		$(BIN).sha256=$(INSTALL_LOCATION)/$(BIN).sha256 \
		$(BIN).sha512=$(INSTALL_LOCATION)/$(BIN).sha512 \
		$(BIN).phash=$(INSTALL_LOCATION)/$(BIN).phash \
		$(BIN).oshash=$(INSTALL_LOCATION)/$(BIN).oshash; \
	fi

package-tar-arm: $(BIN_DIR)/$(BIN)_$(LINUX_OS)_$(ARCHARM) pkgs
	@if [ -f $(PKGS_DIR)/$(BIN)_$(VERSION)_$(ARCHARM).tar ]; then \
		echo "$(PKGS_DIR)/$(BIN)_$(VERSION)_$(ARCHARM).tar already exists"; \
	else \
		$(MAKE) build-symlinks && \
		fpm -s dir -t tar -n $(BIN)_$(VERSION)_$(ARCHARM) -p pkgs -v $(VERSION) -a $(ARCHARM) \
		--license "$(LICENSE)" --description="$(DESCRIPTION)" -m "$(MAINTAINER)" --url "$(URL)" \
		-C $(BIN_DIR)/ $(BIN)_$(LINUX_OS)_$(ARCHARM)=$(INSTALL_LOCATION)/$(BIN) \
		$(BIN).sha1=$(INSTALL_LOCATION)/$(BIN).sha1 \
		$(BIN).md5=$(INSTALL_LOCATION)/$(BIN).md5 \
		$(BIN).sha256=$(INSTALL_LOCATION)/$(BIN).sha256 \
		$(BIN).sha512=$(INSTALL_LOCATION)/$(BIN).sha512 \
		$(BIN).phash=$(INSTALL_LOCATION)/$(BIN).phash \
		$(BIN).oshash=$(INSTALL_LOCATION)/$(BIN).oshash; \
	fi

package-slackware-intel: $(BIN_DIR)/$(BIN)_$(LINUX_OS)_$(ARCHINTEL) pkgs
	@if [ -f $(PKGS_DIR)/$(BIN)-$(VERSION)-$(ARCHINTEL)-1_GG.tgz ]; then \
		echo "$(PKGS_DIR)/$(BIN)-$(VERSION)-$(ARCHINTEL)-1_GG.tgz already exists"; \
	else \
		$(MAKE) build-symlinks && \
		$(MAKE) build-symlinks && \
		$(MAKE) package-slackware-info && \
		cd ./packaging/slackware && \
		sudo cp makepkg /sbin/makepkg && \
		sudo ARCH='$(ARCHINTEL)' VERSION='$(VERSION)' OUTPUT="$$(pwd)/../../pkgs/" ./integrity.SlackBuild ;\
	fi

package-slackware-arm: $(BIN_DIR)/$(BIN)_$(LINUX_OS)_$(ARCHARM) pkgs
	@if [ -f $(PKGS_DIR)/$(BIN)-$(VERSION)-$(ARCHARM)-1_GG.tgz ]; then \
		echo "$(PKGS_DIR)/$(BIN)-$(VERSION)-$(ARCHARM)-1_GG.tgz already exists"; \
	else \
		$(MAKE) build-symlinks && \
		$(MAKE) build-symlinks && \
		$(MAKE) package-slackware-info && \
		cd ./packaging/slackware && \
		sudo cp makepkg /sbin/makepkg && \
		sudo ARCH='$(ARCHARM)' VERSION='$(VERSION)' OUTPUT="$$(pwd)/../../pkgs/" ./integrity.SlackBuild ;\
	fi

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
	go test -v -cover ./pkg/integrity

test-add-linux-attr:
	setfattr -n user.test -v "This is the user.test attribute" data.dat
	setfattr -n test -v "This is the test test attribute" data.dat

test-list-linux-attr:
	getfattr -d data.dat

test-phash:
	./bin/integrity.phash -afv pkg/integrity/testdata/jpeg-sml.jpeg
	@echo "pkg/integrity/testdata/jpeg-sml.jpeg : phash : 8000000000000000 : added"

test-make-data.dat:
	echo 'hello world' > data.dat

test-github-package:
	act push -j package

test-github-test:
	act push -j test

show-version:
	@echo $(VERSION)

git-create-tag:
	git tag -a v$(VERSION) -m "v$(VERSION)"

git-push-tag:
	git push origin v$(VERSION)

git-version-push:
	git add pkg/integrity/version.go
	git commit -m ":bookmark: Updated to version v$(VERSION)"
	git push
	$(MAKE) git-create-tag
	$(MAKE) git-push-tag

record-examples:
	./demos/integrity_generate_casts.sh