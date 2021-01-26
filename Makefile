GOC=go build
FETCHLIBS=go get

BUILDDIR=$(CURDIR)/build

SRCDIR=src/

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
	$(GOC) -o $(OUTPUT)

install:
	mkdir -p $(BINDIR)
	$(INSTALL_BIN) $(OUTPUT) $(BINDIR)/

clean:
	rm -fr $(BUILDDIR)
	rm -f $(OUTPUT)
