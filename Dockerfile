FROM alpine

RUN apk --update add ca-certificates && rm -rf /var/cache/apk/*

RUN apk --update add curl tar \
  && mkdir /ttnsrc \
  && curl -sL -o /ttnsrc/ttn-linux-amd64.tar.gz "https://ttnreleases.blob.core.windows.net/release/src/github.com/TheThingsNetwork/ttn/release/branch/develop/ttn-linux-amd64.tar.gz" \
  && tar -xf /ttnsrc/ttn-linux-amd64.tar.gz -C /ttnsrc \
  && mv /ttnsrc/ttn-linux-amd64 /usr/local/bin/ttn \
  && chmod 755 /usr/local/bin/ttn \
  && rm -rf /ttnsrc \
  && apk del curl tar \
  && rm -rf /var/cache/apk/*

ENTRYPOINT ["/usr/local/bin/ttn"]
