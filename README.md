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

Shorty uses Bolt for persisting all shortened URLs, so no need to setup a
database server. However, this implies that you cannot distribute Shorty onto
multiple nodes.

## Usage

Shorty provides exactly two HTTP endpoints:

1. `http://localhost:3002/shorten?url=<URL>` for shortening a URL. It returns
   the shortened URL in the payload of the response.
1. `http://localhost:3002/s/<SHORT>` is a shortened URL and returns an HTTP 301
   with the location header set to the destination URL.

If you would like to use the HTML UI provided in `assets/`, simply copy the
files included in `assets/` to somewhere reachable by your web server and make
it proxy requests to `/s/` and `/shorten` to Shorty (running on port 3002). See
below for an example Nginx configuration.

## Example nginx configuration

This configuration assumes shorty is running with the `-host` parameter set to
`YOURDOMAIN` and the `index.html` file placed in `/home/makkes/shorty/www/`.

```
server {
    listen              80;
    server_name         YOURDOMAIN;

    access_log /var/log/nginx/shorty_access.log;
    error_log /var/log/nginx/shorty_error.log;

    root /home/makkes/shorty/www/;

    gzip on;
    gzip_proxied any;
    gzip_types text/css application/javascript;

    location ~ /(shorten|s/) {
        client_max_body_size 5g;

        proxy_pass http://127.0.0.1:3002;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwared-For $proxy_add_x_forwarded_for;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

## License

This software is distributed under the BSD 2-Clause License, see
[LICENSE](LICENSE) for more information.
