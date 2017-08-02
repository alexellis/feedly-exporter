FROM golang:1.7.5

WORKDIR /go/src/github.com/alexellis/feedly_exporter
COPY app.go /go/src/github.com/alexellis/feedly_exporter

RUN go get -d -v

RUN go build -o feedly_exporter

FROM alpine:3.5

RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/alexellis/feedly_exporter/feedly_exporter /root/

EXPOSE 9001

ENTRYPOINT ["./feedly_exporter"]
