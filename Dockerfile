FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY esproxy .
ENTRYPOINT ["./esproxy"]

EXPOSE 19200 8080
