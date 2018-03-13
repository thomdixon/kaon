FROM golang

ADD . /go/src/github.com/thomdixon/kaon

RUN go get -u github.com/go-redis/redis
RUN go install github.com/thomdixon/kaon

ENTRYPOINT /go/bin/kaon

EXPOSE 8080