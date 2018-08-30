FROM golang:1.11

WORKDIR /go/src/github.com/mercari/certificate-expiry-monitor-controller

RUN go get github.com/golang/dep/cmd/dep
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -v -vendor-only

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go install -v \
            -ldflags="-w -s" \
            -ldflags "-X main.serviceName=certificate-expiry-monitor-controller" \
            github.com/mercari/certificate-expiry-monitor-controller

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=0 /go/bin/certificate-expiry-monitor-controller /bin/certificate-expiry-monitor-controller

CMD ["/bin/certificate-expiry-monitor-controller"]