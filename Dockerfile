FROM alpine:3.10
RUN apk --update --no-cache add ca-certificates
ADD ./release/ttn-linux-amd64 /usr/local/bin/ttn
RUN chmod 755 /usr/local/bin/ttn
ENTRYPOINT ["/usr/local/bin/ttn"]
