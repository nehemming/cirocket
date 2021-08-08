# Self building docker image
FROM ghcr.io/nehemming/gobuilder:latest as builder
RUN mkdir -p /build
ADD . /build/
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux go test ./... && \
    VERSION=$(svu) && \
    CGO_ENABLED=0 GOOS=linux go build -a -o cirocket -ldflags="-s -w -X main.version=$VERSION" && \
    upx cirocket

# generate clean, final image for end users
FROM alpine:3.13
WORKDIR /opt/app
COPY --from=builder /build/cirocket .
USER 1000:1000
# executable
ENTRYPOINT [ "./cirocket" ]





