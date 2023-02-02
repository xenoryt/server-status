FROM alpine:latest

RUN apk update -U && \
    apk add ffmpeg

COPY ./stream-status /bin/stream-status

ENTRYPOINT stream-status
