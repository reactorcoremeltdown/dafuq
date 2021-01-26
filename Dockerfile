FROM golang:buster

COPY . /data
ENV GOBIN=/usr/local/bin
WORKDIR /data

CMD go get && go build -o /opt/dafuq
