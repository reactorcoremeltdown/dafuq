all: dafuq

dafuq: Makefile Dockerfile main.go
	test -d /opt/apps/dafuq || mkdir -p /opt/apps/dafuq
	podman build -t repo.rcmd.space/dafuq:latest .
	podman push repo.rcmd.space/dafuq:latest

release: server linux_amd64 linux_arm64 linux_arm

server:
	systemctl stop wtfd
	podman run --env=GOOS=linux --env=GOARCH=amd64 -v /opt/apps/dafuq:/opt/apps/dafuq repo.rcmd.space/dafuq:latest sh -c 'unset GOBIN && go get && go build -o /opt/dafuq-linux_amd64; cp /opt/dafuq-linux_amd64 /opt/apps/dafuq/'
	systemctl start wtfd

linux_amd64:
	test -d /opt/apps/dafuq/releases/${DRONE_TAG} || mkdir -p /opt/apps/dafuq/releases/${DRONE_TAG}
#	podman run --env=GOOS=linux --env=GOARCH=amd64 -v /opt/apps/dafuq:/opt/apps/dafuq repo.rcmd.space/dafuq:latest sh -c 'unset GOBIN && go get && go build -o /opt/dafuq-linux_amd64; cp /opt/dafuq-linux_amd64 /opt/apps/dafuq/'

linux_arm64:
	test -d /opt/apps/dafuq/releases/${DRONE_TAG} || mkdir -p /opt/apps/dafuq/releases/${DRONE_TAG}
	podman run --env=GOOS=linux --env=GOARCH=arm64 -v /opt/apps/dafuq:/opt/apps/dafuq repo.rcmd.space/dafuq:latest sh -c 'unset GOBIN && go get && go build -o /opt/dafuq-linux_arm64; cp /opt/dafuq-linux_arm64 /opt/apps/dafuq/'

linux_arm:
	test -d /opt/apps/dafuq/releases/${DRONE_TAG} || mkdir -p /opt/apps/dafuq/releases/${DRONE_TAG}
	podman run --env=GOOS=linux --env=GOARCH=arm -v /opt/apps/dafuq:/opt/apps/dafuq repo.rcmd.space/dafuq:latest sh -c 'unset GOBIN && go get && go build -o /opt/dafuq-linux_arm; cp /opt/dafuq-linux_arm /opt/apps/dafuq/'

# darwin_amd64:
# 	podman run --env=GOOS=darwin --env=GOARCH=amd64 -v /opt/apps/dafuq:/opt/apps/dafuq repo.rcmd.space/dafuq:latest sh -c 'unset GOBIN && go get && go build -o /opt/dafuq-darwin_amd64; cp /opt/dafuq-darwin_amd64 /opt/apps/dafuq/'
# 
# darwin_arm64:
# 	podman run --env=GOOS=darwin --env=GOARCH=arm64 -v /opt/apps/dafuq:/opt/apps/dafuq repo.rcmd.space/dafuq:latest sh -c 'unset GOBIN && go get && go build -o /opt/dafuq-darwin_arm64; cp /opt/dafuq-darwin_arm64 /opt/apps/dafuq/'
