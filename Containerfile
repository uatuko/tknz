FROM --platform=$BUILDPLATFORM golang:1-trixie AS builder

ARG TARGETOS
ARG TARGETARCH

COPY go.mod go.sum main.go /tmp/buildroot/
COPY .dist/ /opt/package/.dist/
COPY internal/ /tmp/buildroot/internal/
COPY pb/ /tmp/buildroot/pb/

RUN \
	cd /tmp/buildroot \
	&& \
	CGO_ENABLED=0 \
	GOARCH=${TARGETARCH} \
	GOOS=${TARGETOS} \
	go build \
		-o /opt/package/bin/tknz \
		main.go


FROM debian:trixie-slim

RUN \
	apt-get update \
	&& \
	apt-get install -y ca-certificates \
	&& \
	apt-get clean \
	&& \
	rm -r /var/lib/apt/lists/*

COPY --from=builder /opt/package /opt/package

WORKDIR /opt/package
ENTRYPOINT [ "bin/tknz" ]
EXPOSE 8080
