FROM golang:1.23-alpine as build-stage

WORKDIR /tmp/build

COPY . .

# Build the project
RUN go build

FROM alpine:3

LABEL name "Xinny Gateway"
LABEL maintainer "Xinny Developers <xinnydev@frutbits.org>"

WORKDIR /app

# Install needed deps
RUN apk add --no-cache vips tini

COPY --from=build-stage /tmp/build/gateway gateway

ENTRYPOINT ["tini", "--"]
CMD ["/app/gateway"]