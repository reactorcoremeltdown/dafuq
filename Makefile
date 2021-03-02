all: dafuq artifacts

dafuq: Makefile Dockerfile main.go
	podman build -t dafuq:latest .
	systemctl stop wtfd.service
	podman run -v /opt/apps/dafuq:/opt/apps/dafuq dafuq:latest cp /opt/dafuq /opt/apps/dafuq/
	systemctl start wtfd.service

artifacts: linux_amd64 linux_arm64

linux_amd64:
	podman build --build-arg=GOOS=linux --build-arg=GOARCH=amd64 -t dafuq:latest-linux-amd64 .
	podman run -v /opt/apps/dafuq:/opt/apps/dafuq dafuq:latest-linux-amd64 cp /opt/dafuq /opt/apps/dafuq/dafuq-linux_amd64

linux_arm64:
	podman build --build-arg=GOOS=linux --build-arg=GOARCH=arm64 -t dafuq:latest-linux-arm64 .
	podman run -v /opt/apps/dafuq:/opt/apps/dafuq dafuq:latest-linux-arm64 cp /opt/dafuq /opt/apps/dafuq/dafuq-linux_arm64
