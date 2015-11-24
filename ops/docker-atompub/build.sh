#!/usr/bin/env sh

vagrant docker-run builder -- \
  /go/src/github.com/ironcamel/go.atompub/ops/docker-atompub/container-build.sh
