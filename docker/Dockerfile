FROM alpine
RUN apk --update add git less openssh && rm -rf /var/lib/apt/lists/* && rm /var/cache/apk/*
COPY gitdm-sync /usr/bin/
CMD lf-gitdm-sync
