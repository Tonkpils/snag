FROM golang:1.5

CMD snag
ENV	PATH	$GOPATH:$PATH

WORKDIR $GOPATH/src/github.com/Tonkpils/snag
ADD . $GOPATH/src/github.com/Tonkpils/snag

RUN go get -t 
RUN go build -ldflags "-X main.runningDocker=true" -o $GOPATH/bin/snag

WORKDIR /