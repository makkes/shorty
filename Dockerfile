ARG GO_VERSION=1.25
ARG XX_VERSION=1.6.1

FROM --platform=$BUILDPLATFORM tonistiigi/xx:${XX_VERSION} AS xx
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS builder

COPY --from=xx / /

# We need git for Go to figure out version information to be put into the binary
RUN apk add git
RUN git --version

WORKDIR /shorty
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

# IMPORTANT: always copy the whole repository tree so that the .git directory is available to determine the git revision etc.
COPY . .

ARG TARGETPLATFORM
ENV CGO_ENABLED=0
# important to pass -buildvcs=true so the binary has version information available
RUN xx-go build -trimpath -a -o shorty -ldflags="-w -s" -buildvcs=true .

FROM gcr.io/distroless/static-debian12:nonroot@sha256:627d6c5a23ad24e6bdff827f16c7b60e0289029b0c79e9f7ccd54ae3279fb45f

COPY --from=builder /shorty/shorty .
COPY --from=builder /shorty/assets assets

CMD ["./shorty"]
