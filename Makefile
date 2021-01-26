all: dafuq

dafuq: Makefile Dockerfile main.go
	docker build -t dafuq:latest .

install:
	mkdir -p $(BINDIR)
	$(INSTALL_BIN) $(OUTPUT) $(BINDIR)/

clean:
	rm -fr $(BUILDDIR)
	rm -f $(OUTPUT)
