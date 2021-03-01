FROM golang:buster

COPY . /data
ENV GOBIN=/usr/local/bin
WORKDIR /data
RUN go mod init && go get && go build -o /opt/dafuq

CMD /bin/true
