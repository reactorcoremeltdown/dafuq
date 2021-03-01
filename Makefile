all: dafuq

dafuq: Makefile Dockerfile main.go
	podman build -t dafuq:latest .
	systemctl stop wtfd.service
	podman run -v /opt/apps/dafuq:/opt/apps/dafuq dafuq:latest cp /opt/dafuq /opt/apps/dafuq/
	systemctl start wtfd.service
