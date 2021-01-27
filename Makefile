all: dafuq

dafuq: Makefile Dockerfile main.go
	docker build -t dafuq:latest .
	docker run -v /opt/apps/dafuq:/opt/apps/dafuq dafuq:latest cp /opt/dafuq /opt/apps/dafuq/
