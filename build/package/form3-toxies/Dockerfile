FROM golang:1.16.3-alpine AS build-env

ARG APPNAME
ENV GO111MODULE=auto
ENV SRCPATH $GOPATH/src/github.com/form3tech-oss/$APPNAME

COPY ./ $SRCPATH

RUN go install github.com/form3tech-oss/$APPNAME/cmd/$APPNAME

FROM alpine

ARG APPNAME
WORKDIR /app

COPY --from=build-env /go/bin/$APPNAME /app/

RUN ip -4 route list match 0/0 | awk '{print $3 "host.docker.internal"}' >> /etc/hosts

EXPOSE 8474
ENTRYPOINT ["./form3-toxies"]
CMD ["-host=0.0.0.0"]

ENV LOG_LEVEL=info
