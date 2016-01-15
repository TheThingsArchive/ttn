# The container will assume to have a $BROKER env var defined

# Go, because go is life
FROM golang:latest

# Dependencies, everything on master
RUN go get "github.com/TheThingsNetwork/ttn/..."

# Actual files to build
RUN mkdir ~/TheThingsNetwork
ADD . ~/TheThingsNetwork
WORKDIR ~/TheThingsNetwork

# Expose the port on which the adapters will listen to
EXPOSE 3000/tcp
EXPOSE 4000/tcp

# Build & Launch
RUN go build .
CMD ./broker --routers-port 3000 --handlers-port 4000
