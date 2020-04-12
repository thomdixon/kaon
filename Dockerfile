FROM golang

ADD . /go/src/github.com/thomdixon/kaon

WORKDIR /go/src/github.com/thomdixon/kaon
RUN go mod download
RUN go install github.com/thomdixon/kaon

ENTRYPOINT /go/bin/kaon

EXPOSE 8080