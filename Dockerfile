FROM alpine:latest

WORKDIR /app
ARG TARGETPLATFORM

RUN apk add --no-cache tzdata ca-certificates

ENV TZ=UTC

COPY $TARGETPLATFORM/watchducker /app
COPY docker/entrypoint.sh /usr/local/bin/entrypoint.sh
COPY push.yaml.example /app

RUN chmod +x /app/watchducker /usr/local/bin/entrypoint.sh && \
    ln -s /app/watchducker /usr/local/bin/watchducker

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
CMD ["watchducker"]
