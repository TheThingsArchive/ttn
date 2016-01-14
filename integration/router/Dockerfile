# The container will assume to have a $BROKER env var defined

# Go, because go is life
FROM golang:latest

# Dependencies, everything on master
RUN go get "github.com/TheThingsNetwork/core/..."

# Actual files to build
RUN mkdir ~/TheThingsNetwork
ADD . ~/TheThingsNetwork
WORKDIR ~/TheThingsNetwork

# Expose the port on which the gateway adapter will listen to
EXPOSE 33000/udp

# Build & Launch
RUN go build .
CMD ./router --brokers $BROKERS --udp-port 33000
