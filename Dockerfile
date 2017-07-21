FROM golang:latest

ADD . /go/src/github.com/spencercdixon/aws-cli-server

RUN go install /go/src/github.com/spencercdixon/aws-cli-server

ENTRYPOINT /go/bin/aws-cli-server

EXPOSE 4000

