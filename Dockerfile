FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -o /out/watchducker .

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache tzdata ca-certificates

COPY --from=builder /out/watchducker /app/watchducker
COPY push.yaml.example /app

RUN chmod +x /app/watchducker && \
    ln -s /app/watchducker /usr/local/bin/watchducker

CMD ["watchducker"]
