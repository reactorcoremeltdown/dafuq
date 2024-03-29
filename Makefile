DRONE_TAG=$(shell git describe --tags --abbrev=0)
DRONE_COMMIT_ID=$(shell git log --format="%H" -n 1)
DRONE_BUILD_DATE=$(shell date "+%Y-%m-%d")
DRONE_VERSION_LINE := version: $(DRONE_TAG), build date: $(DRONE_BUILD_DATE), commit ID: $(DRONE_COMMIT_ID)
GOLANG_VERSION="1.20.5-buster"

all: dafuq

dafuq: Makefile Dockerfile main.go
	test -d /opt/apps/dafuq || mkdir -p /opt/apps/dafuq
	test -d /var/cache/golang || mkdir -p /var/cache/golang
	systemctl stop wtfd
	podman run --env=GOOS=linux \
		--env=GOARCH=amd64 \
		--env=GOPATH=/var/cache/golang \
		-v $(shell pwd):/app/dafuq \
		-v /var/cache/golang:/var/cache/golang \
		-v /opt/apps/dafuq:/opt/apps/dafuq \
		golang:latest sh -c "unset GOBIN && cd /app/dafuq && go get && go build -ldflags='-X main.Version=${DRONE_TAG} -X main.CommitID=${DRONE_COMMIT_ID} -X main.BuildDate=${DRONE_BUILD_DATE}' -buildvcs=false -o /opt/dafuq-linux_amd64; cp /opt/dafuq-linux_amd64 /opt/apps/dafuq/"
	systemctl start wtfd

release: linux_amd64 linux_arm64 linux_arm

linux_amd64:
	test -d /opt/apps/dafuq/releases/${DRONE_TAG} || mkdir -p /opt/apps/dafuq/releases/${DRONE_TAG}
	podman run --env=GOOS=linux \
		--env=GOARCH=amd64 \
		--env=GOPATH=/var/cache/golang \
		-v $(shell pwd):/app/dafuq \
		-v /var/cache/golang:/var/cache/golang \
		-v /opt/apps/dafuq/releases/${DRONE_TAG}:/opt/apps/dafuq \
		golang:${GOLANG_VERSION} sh -c "unset GOBIN && cd /app/dafuq && go get && go build -ldflags='-X main.Version=${DRONE_TAG} -X main.CommitID=${DRONE_COMMIT_ID} -X main.BuildDate=${DRONE_BUILD_DATE}' -buildvcs=false -o /opt/dafuq-linux_amd64; cp /opt/dafuq-linux_amd64 /opt/apps/dafuq/"

linux_arm64:
	test -d /opt/apps/dafuq/releases/${DRONE_TAG} || mkdir -p /opt/apps/dafuq/releases/${DRONE_TAG}
	podman run --env=GOOS=linux \
		--env=GOARCH=arm64 \
		--env=GOPATH=/var/cache/golang \
		-v $(shell pwd):/app/dafuq \
		-v /var/cache/golang:/var/cache/golang \
		-v /opt/apps/dafuq/releases/${DRONE_TAG}:/opt/apps/dafuq \
		golang:${GOLANG_VERSION} sh -c "unset GOBIN && cd /app/dafuq && go get && go build -ldflags='-X main.Version=${DRONE_TAG} -X main.CommitID=${DRONE_COMMIT_ID} -X main.BuildDate=${DRONE_BUILD_DATE}' -buildvcs=false -o /opt/dafuq-linux_arm64; cp /opt/dafuq-linux_arm64 /opt/apps/dafuq/"

linux_arm:
	test -d /opt/apps/dafuq/releases/${DRONE_TAG} || mkdir -p /opt/apps/dafuq/releases/${DRONE_TAG}
	podman run --env=GOOS=linux \
		--env=GOARCH=arm \
		--env=GOPATH=/var/cache/golang \
		-v $(shell pwd):/app/dafuq \
		-v /var/cache/golang:/var/cache/golang \
		-v /opt/apps/dafuq/releases/${DRONE_TAG}:/opt/apps/dafuq \
		golang:${GOLANG_VERSION} sh -c "unset GOBIN && cd /app/dafuq && go get && go build -ldflags='-X main.Version=${DRONE_TAG} -X main.CommitID=${DRONE_COMMIT_ID} -X main.BuildDate=${DRONE_BUILD_DATE}' -buildvcs=false -o /opt/dafuq-linux_arm; cp /opt/dafuq-linux_arm /opt/apps/dafuq/"

# darwin_amd64:
# 	podman run --env=GOOS=darwin --env=GOARCH=amd64 -v /opt/apps/dafuq:/opt/apps/dafuq repo.rcmd.space/dafuq:latest sh -c 'unset GOBIN && go get && go build -o /opt/dafuq-darwin_amd64; cp /opt/dafuq-darwin_amd64 /opt/apps/dafuq/'
# 
# darwin_arm64:
# 	podman run --env=GOOS=darwin --env=GOARCH=arm64 -v /opt/apps/dafuq:/opt/apps/dafuq repo.rcmd.space/dafuq:latest sh -c 'unset GOBIN && go get && go build -o /opt/dafuq-darwin_arm64; cp /opt/dafuq-darwin_arm64 /opt/apps/dafuq/'
