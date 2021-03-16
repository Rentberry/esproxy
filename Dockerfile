FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY esproxy_linux_amd64 .
ENTRYPOINT ["./esproxy"]

EXPOSE 19200 8080
