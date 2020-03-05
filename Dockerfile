FROM golang:alpine

COPY . /print_queue
WORKDIR /print_queue/print_queue_server

RUN apk add git
RUN go get
EXPOSE 12345
ENTRYPOINT go run ./print_queue.server.go
