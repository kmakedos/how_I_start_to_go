FROM alpine:3.14
RUN apk --no-cache add curl
COPY weather /
ENTRYPOINT ["/weather"]
