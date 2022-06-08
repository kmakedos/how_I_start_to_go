FROM alpine:3.14
COPY weather /
ENTRYPOINT ["/weather"]
