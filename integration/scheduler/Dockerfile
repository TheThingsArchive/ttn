# The container will assume to have a $ROUTERS env var defined

# Go, because go is life
FROM golang:latest

# Dependencies, everything on master
RUN go get "github.com/TheThingsNetwork/ttn/..."

# Actual files to build
RUN mkdir ~/TheThingsNetwork
ADD . ~/TheThingsNetwork
WORKDIR ~/TheThingsNetwork

# Build & Launch
RUN go build -o scheduler .
CMD ./scheduler --schedule "./schedule.json" --delay "1s" --routers $ROUTERS
