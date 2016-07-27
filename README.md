# Shorty

A very simple URL shortener with an even simpler UI.

# Installation

To install Shorty, run 

```
go get github.com/makkes/shorty
```

or grab the binary of the [most current
release](https://github.com/makkes/shorty/releases).

# Running

To start Shorty, simply call `shorty` (assuming that `$GOPATH/bin` is on your
`$PATH`), passing the parameters fit to your environment (see below).

# Configuration/Persistence

The startup configuration of Shorty is provided via command-line parameters.
Type `shorty -h` to get a list of all parameters.

Shorty uses Bolt for persisting all shortened URLs, so no need to setup a
database server. However, this implies that you cannot distribute Shorty onto
multiple nodes.

# Usage

Shorty provides exactly two HTTP endpoints:

1. `http://localhost:3002/shorten?url=<URL>` for shortening a URL. It returns
   the shortened URL in the payload of the response.
1. `http://localhost:3002/s/<SHORT>` is a shortened URL and returns an HTTP 301
   with the location header set to the destination URL.

If you would like to use the HTML UI provided in `assets/html/`, simply copy the
`index.html` file to somewhere reachable by your web server and make it proxy
requests to `/s/` and `/shorten` to Shorty (running on port 3002).

# License

This software is distributed under the BSD 2-Clause License, see
[LICENSE](LICENSE) for more information.
