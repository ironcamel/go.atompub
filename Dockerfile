FROM phusion/baseimage:latest
MAINTAINER will@tilt.com

RUN mkdir -p /opt/go.atompub
ADD ./bin/go.atompub /opt/go.atompub
ADD ./ops/docker-atompub/runit /etc/service/atompub

ENTRYPOINT ["/sbin/my_init", "--"]

EXPOSE 8000
