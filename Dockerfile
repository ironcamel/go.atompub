FROM phusion/baseimage:latest
MAINTAINER will@tilt.com

# GO ==========================================================================
# Stolen from: # https://github.com/docker-library/golang/blob/master/1.5/Dockerfile

# gcc for cgo
RUN apt-get update && apt-get install -y --no-install-recommends \
    g++ \
    gcc \
    libc6-dev \
    make \
  && rm -rf /var/lib/apt/lists/*

ENV GOLANG_VERSION 1.5.1
ENV GOLANG_DOWNLOAD_URL https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz
ENV GOLANG_DOWNLOAD_SHA1 46eecd290d8803887dec718c691cc243f2175fe0

RUN curl -fsSL "$GOLANG_DOWNLOAD_URL" -o golang.tar.gz \
  && echo "$GOLANG_DOWNLOAD_SHA1  golang.tar.gz" | sha1sum -c - \
  && tar -C /usr/local -xzf golang.tar.gz \
  && rm golang.tar.gz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
WORKDIR $GOPATH

# END GO ======================================================================

# GIT
RUN apt-get -y update && apt-get -y install git

RUN mkdir -p /go/src/github.com/ironcamel/go.atompub
COPY ./* /go/src/github.com/ironcamel/go.atompub/

RUN cd /go/src/github.com/ironcamel/go.atompub \
  && go get \
  && go install

ADD ./ops/docker-atompub/runit /etc/service/atompub

# Using ENTRYPOINT instead of CMD to overwrite plenv's ENTRYPOINT
ENTRYPOINT ["/sbin/my_init", "--"]

EXPOSE 8000
