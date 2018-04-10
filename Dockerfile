FROM golang:alpine as build

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh gcc musl-dev
ENV GOROOT=/usr/local/go
RUN go get -v github.com/onrik/ethrpc
RUN go get -v github.com/gorilla/mux
RUN go get -v github.com/kardianos/service
COPY . /usr/local/go/src/github.com/decosblockchain/audittrail-server
WORKDIR /usr/local/go/src/github.com/decosblockchain/audittrail-server
RUN go get -v ./...
RUN go build


FROM alpine
WORKDIR /app
RUN cd /app
COPY --from=build /usr/local/go/src/github.com/decosblockchain/audittrail-server/audittrail-server /app/bin/audittrail-server

RUN mkdir -p /app/bin/data

EXPOSE 3000
WORKDIR /app/bin

CMD ["./audittrail-server"]