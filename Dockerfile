FROM golang:1.8.5-alpine as builder

WORKDIR /go/src/github.com/makkes/shorty
RUN apk add --no-cache git
COPY . .
RUN go get
RUN go build

FROM alpine:latest

RUN apk update && apk add ca-certificates

WORKDIR /root
COPY --from=builder /go/src/github.com/makkes/shorty/shorty .
COPY --from=builder /go/src/github.com/makkes/shorty/assets assets

CMD ["./shorty"]
