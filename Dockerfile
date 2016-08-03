FROM alpine
RUN apk --update add ca-certificates && rm -rf /var/cache/apk/*
ADD ./release/ttn-linux-amd64 /usr/local/bin/ttn
RUN chmod 755 /usr/local/bin/ttn
ENTRYPOINT ["/usr/local/bin/ttn"]
