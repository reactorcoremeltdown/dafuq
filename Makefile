GOC := /usr/bin/go build
FETCHLIBS=/usr/bin/go get

BUILDDIR=$(CURDIR)/build
GOBINDIR=$(BUILDDIR)/bin
GOPATHDIR=$(BUILDDIR)/golibs

INSTALL=install
INSTALL_BIN=$(INSTALL) -m755
INSTALL_LIB=$(INSTALL) -m644
INSTALL_CONF=$(INSTALL) -m400

PREFIX?=$(DESTDIR)/usr
BINDIR?=$(PREFIX)/bin

all: dafuq

dafuq: Makefile main.go
	mkdir -p $(GOPATHDIR) && \
	mkdir -p $(GOBINDIR) && \
	export GOPATH=$(GOPATHDIR) && \
	export GOBIN=$(GOBINDIR) && \
	$(FETCHLIBS) && \
	$(GOC) -o dafuq

install:
	pwd && ls -la
	mkdir -p $(BINDIR)
	$(INSTALL_BIN) dafuq $(BINDIR)/

package:
	DRONE_COMMIT_ID := ${DRONE_COMMIT_ID}
	DRONE_TAG := ${DRONE_TAG}
	wget -O- https://raw.githubusercontent.com/rcmd-funkhaus/debrewery/master/debrew.sh | bash -
