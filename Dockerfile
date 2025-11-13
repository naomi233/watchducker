FROM alpine:latest

WORKDIR /app
ARG TARGETPLATFORM

RUN apk add --no-cache tzdata

COPY $TARGETPLATFORM/watchducker /app
COPY push.yaml.example /app

RUN chmod +x /app/watchducker && \
    ln -s /app/watchducker /usr/local/bin/watchducker

CMD ["watchducker"]
