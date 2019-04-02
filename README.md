# Shorty

[![Build Status](https://travis-ci.org/makkes/shorty.svg?branch=master)](https://travis-ci.org/makkes/shorty)

A very simple URL shortener with an even simpler UI.

## Installation

To install Shorty, you have 3 options: 

1. Install using Go:
   ```
   go get github.com/makkes/shorty
   ```
   Then you can just run `shorty` (see below for runtime parameters).
2. Grab the binary of the [most current
   release](https://github.com/makkes/shorty/releases).
3. Get the Docker image:
   ```
   docker pull makkes/shorty:VERSION
   ```
   `VERSION` is either `latest` or a release number such as `v1.1.0`.

## Configuration/Persistence

The startup configuration of Shorty is provided via environment variables:

|Variable|Description|Default
|---|---|---
|`LISTEN_HOST`|The IP address/hostname to listen on|`localhost`
|`LISTEN_PORT`|The port to listen on|`3002`
|`SERVE_HOST`|The host used by users to reach Shorty|`localhost`
|`SERVE_PROTOCOL`|One of `http` or `https`|`https`
|`BACKEND`|The persistence backend to use, one of `bolt`, `dynamodb`|`bolt`

Shorty provides two persistence mechanisms: A Bolt database, persisting all data
in a single database file and a DynamoDB backend, storing all data in an AWS
DynamoDB table.

### Bolt Backend Configuration

|Variable|Description|Default
|---|---|---
|`DB_DIR`|The directory used to store Shorty's database files|the current directory

When you choose the Bolt backend, you don't need to setup a database server.
However, this implies that you cannot distribute Shorty onto multiple nodes.

### DynamoDB Backend Configuration

The DynamoDB backend is configured via standard AWS environment variables; see
https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html?shortFooter=true
for an explanation.

## License

This software is distributed under the BSD 2-Clause License, see
[LICENSE](LICENSE) for more information.

