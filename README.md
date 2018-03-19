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
|DB_DIR|The directory used to store Shorty's database files|the current directory

Shorty uses Bolt for persisting all shortened URLs, so no need to setup a
database server. However, this implies that you cannot distribute Shorty onto
multiple nodes.

## Running Docker image

There are Docker images available at https://hub.docker.com/r/makkes/shorty/.

## License

This software is distributed under the BSD 2-Clause License, see
[LICENSE](LICENSE) for more information.

