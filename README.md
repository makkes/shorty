# Shorty

[![Build Status](https://travis-ci.org/makkes/shorty.svg?branch=master)](https://travis-ci.org/makkes/shorty)

A very simple URL shortener with an even simpler UI.

## Installation

To install Shorty, run 

```
go get github.com/makkes/shorty
```

or grab the binary of the [most current
release](https://github.com/makkes/shorty/releases).

## Running

To start Shorty, simply call `shorty` (assuming that `$GOPATH/bin` is on your
`$PATH`), passing the parameters fit to your environment (see below).

## Configuration/Persistence

The startup configuration of Shorty is provided via environment variables:

|Variable|Description|Default
|---|---|---
|LISTEN_HOST|The IP address/hostname to listen on|localhost
|LISTEN_PORT|The port to listen on|3002
|SERVE_HOST|The host used by users to reach Shorty|localhost
|SERVE_PROTOCOL|One of 'http' or 'https'|https
|BACKEND|The persistence backend to use, one of 'bolt', 'dynamodb'|bolt

Shorty provides two persistence mechanisms: A Bolt database, persisting all data
in a single database file and a DynamoDB backend, storing all data in an AWS
DynamoDB table.

### Bolt Backend Configuration

|Variable|Description|Default
|DB_DIR|The directory used to store Shorty's database files|the current directory

When you choose the Bolt backend, you don't need to setup a database server.
However, this implies that you cannot distribute Shorty onto multiple nodes.

### DynamoDB Backend Configuration

The DynamoDB backend is configured via standard AWS environment variables; see
https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html?shortFooter=true
for an explanation.

## Running Docker image

There are Docker images available at https://hub.docker.com/r/makkes/shorty/.

## License

This software is distributed under the BSD 2-Clause License, see
[LICENSE](LICENSE) for more information.

