FROM golang:1.12


COPY . /go/src/device-repository
WORKDIR /go/src/device-repository

ENV GO111MODULE=on

RUN go build

EXPOSE 8080

CMD ./device-repository