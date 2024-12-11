GOC := /usr/bin/go build
FETCHLIBS=/usr/bin/go get -v

BUILDDIR=$(CURDIR)/build
GOBINDIR=$(BUILDDIR)/bin
GOPATHDIR=$(BUILDDIR)/golibs

SRCDIR=$(CURDIR)/src/dafuq
CONFSRCDIR=$(CURDIR)/src/conf/etc/dafuq

INSTALL=install
INSTALL_BIN=$(INSTALL) -m755
INSTALL_CONF=$(INSTALL) -m644
INSTALL_SEC=$(INSTALL) -m400

PREFIX?=$(DESTDIR)/usr
BINDIR?=$(PREFIX)/bin
CONFDIR?=$(DESTDIR)/etc/dafuq
SYSTEMDCONFDIR?=$(DESTDIR)/etc/systemd/system
SYSTEMLOGDIR?=$(DESTDIR)/var/log/dafuq

all: dafuq

dafuq: Makefile src/dafuq/main.go
	mkdir -p $(GOPATHDIR) && \
	mkdir -p $(GOBINDIR) && \
	export GOPATH=$(GOPATHDIR) && \
	export GOBIN=$(GOBINDIR) && \
	cd $(SRCDIR) && \
	$(FETCHLIBS) && \
	$(GOC) -o dafuq

install:
	pwd && ls -la
	mkdir -p $(BINDIR)
	$(INSTALL_BIN) $(SRCDIR)/dafuq $(BINDIR)/
	mkdir -p $(CONFDIR)/configs $(CONFDIR)/notifiers $(CONFDIR)/plugins
	$(INSTALL_CONF) $(CONFSRCDIR)/dafuq.ini $(CONFDIR)/
	$(INSTALL_CONF) $(CONFSRCDIR)/configs/checkfile.ini $(CONFDIR)/configs/
	$(INSTALL_BIN) $(CONFSRCDIR)/notifiers/log $(CONFDIR)/notifiers
	$(INSTALL_BIN) $(CONFSRCDIR)/plugins/* $(CONFDIR)/plugins/
	mkdir -p $(SYSTEMLOGDIR)

package:
	wget -O- https://raw.githubusercontent.com/rcmd-funkhaus/debrewery/master/debrew.sh | bash -
