ARG app_image
ARG tmpl_image

FROM ${app_image} AS app
FROM ${tmpl_image} AS tmpl


FROM --platform=$BUILDPLATFORM golang:1-trixie AS builder

ARG TARGETOS
ARG TARGETARCH

COPY go.mod go.sum main.go /tmp/build/
COPY internal/ /tmp/build/internal/
COPY pb/ /tmp/build/pb/

RUN \
	cd /tmp/build \
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
COPY --from=app / /opt/package/.static/app/
COPY --from=tmpl / /opt/package/.static/tmpl/

WORKDIR /opt/package
ENTRYPOINT [ "bin/tknz" ]
EXPOSE 8080
