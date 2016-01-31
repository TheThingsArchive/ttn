FROM alpine

RUN apk --update add ca-certificates && rm -rf /var/cache/apk/*

RUN apk --update add curl tar \
  && mkdir /ttnsrc \
  && curl -sL -o /ttnsrc/broker-linux-amd64.tar.gz "https://thethingsnetwork.blob.core.windows.net/release/src/github.com/TheThingsNetwork/ttn/release/branch/develop/broker-linux-amd64.tar.gz" \
  && tar -xf /ttnsrc/broker-linux-amd64.tar.gz -C /ttnsrc \
  && mv /ttnsrc/broker-linux-amd64 /usr/local/bin/broker \
  && chmod 755 /usr/local/bin/broker \
  && rm -rf /ttnsrc \
  && apk del curl tar \
  && rm -rf /var/cache/apk/*

ENTRYPOINT ["/usr/local/bin/broker"]
