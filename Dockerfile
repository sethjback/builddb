FROM golang:1.11 as Builder

RUN mkdir -p /go/src/github.com/hexly/noptics/test-utils/buildb
ADD . /go/src/github.com/hexly/noptics/test-utils/buildb

WORKDIR /go/src/github.com/hexly/noptics/test-utils/buildb

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o buildb

FROM alpine:3.9

RUN apk add --no-cache curl bash ca-certificates

COPY --from=builder /go/src/github.com/hexly/noptics/test-utils/buildb/buildb /buildb

CMD ["/buildb"]