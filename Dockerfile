FROM golang:latest

COPY . /data
ENV GOBIN=/usr/local/bin
WORKDIR /data
RUN go get && go build -o /opt/dafuq

CMD /bin/true
