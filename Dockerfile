FROM golang:1.20-alpine as build-stage

WORKDIR /tmp/build

COPY . .

# Install needed deps
RUN apk add --no-cache libc-dev vips-dev gcc g++ make

# Build the project
RUN go build

FROM alpine:3

LABEL name "Xinny Gateway"
LABEL maintainer "Xinny Developers <xinnydev@frutbits.org>"

WORKDIR /app

COPY --from=build-stage /tmp/build/gateway gateway

ENTRYPOINT ["tini", "--"]
CMD ["/app/gateway"]