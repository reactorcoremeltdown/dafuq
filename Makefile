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
	$(GOC)

install:
	mkdir -p $(BINDIR)
	$(INSTALL_BIN) dafuq $(BINDIR)/
