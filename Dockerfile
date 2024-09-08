FROM golang:1.22.2 AS builder

ENV GO111MODULE=on
WORKDIR /go/release
ADD . .
RUN set -x \
    && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -buildid=" -o usdtmore ./main

FROM archlinux:latest

ENV TZ=Asia/Shanghai

COPY --from=builder /go/release/usdtmore /runtime/usdtmore

ADD ./templates /runtime/templates
ADD ./static /runtime/static

EXPOSE 6080
CMD ["/runtime/usdtmore"]