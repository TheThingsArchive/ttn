# Should be placed in root folder til ttn isn't available on master

FROM golang:latest

RUN go get github.com/apex/log
RUN go get github.com/brocaar/lorawan
RUN go get github.com/jacobsa/crypto/cmac
RUN mkdir -p /go/src/github.com/TheThingsNetwork/ttn
ADD . /go/src/github.com/TheThingsNetwork/ttn

RUN go build -o listener github.com/TheThingsNetwork/ttn/integration/udp_debugger
CMD ./listener
