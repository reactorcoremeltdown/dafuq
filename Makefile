all: dafuq

dafuq: Makefile Dockerfile main.go
	podman build -t dafuq:latest .
	podman run -v /opt/apps/dafuq:/opt/apps/dafuq dafuq:latest cp /opt/dafuq /opt/apps/dafuq/
	systemctl restart wtfd.service
