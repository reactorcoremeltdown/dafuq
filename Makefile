all: dafuq

dafuq: Makefile Dockerfile main.go
	docker build -t dafuq:latest .
	install -d /opt/dafuq
