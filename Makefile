all: dafuq artifacts

dafuq: Makefile Dockerfile main.go
	podman build -t repo.rcmd.space/dafuq:latest .
	podman push repo.rcmd.space/dafuq:latest
	systemctl stop wtfd.service
	podman run -v /opt/apps/dafuq:/opt/apps/dafuq repo.rcmd.space/dafuq:latest cp /opt/dafuq /opt/apps/dafuq/
	systemctl start wtfd.service

artifacts: linux_amd64 linux_arm64

linux_amd64:
	podman run --env=GOOS=linux --env=GOARCH=amd64 -v /opt/apps/dafuq:/opt/apps/dafuq repo.rcmd.space/dafuq:latest sh -c 'go get && go build -o /opt/dafuq; cp /opt/dafuq /opt/apps/dafuq/dafuq-linux_amd64'

linux_arm64:
	podman run --env=GOOS=linux --env=GOARCH=arm64 -v /opt/apps/dafuq:/opt/apps/dafuq repo.rcmd.space/dafuq:latest sh -c 'go get && go build -o /opt/dafuq; cp /opt/dafuq /opt/apps/dafuq/dafuq-linux_arm64'
