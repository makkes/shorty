FROM debian:latest

COPY shorty /
COPY assets /assets

CMD ["/shorty"]
