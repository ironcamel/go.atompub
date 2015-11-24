## Getting Started

Install go.atompub:

    go get github.com/ironcamel/go.atompub

Then run the server:

    go run atompub.go

To create a new entry:

```bash
curl -d '
    <entry>
        <title>allo</title>
        <content type="text">{"foo":"bar"}</content>
    </entry>' http://localhost:8000/feeds/widgets
```

That adds a new entry to a feed titled widgets.
If that feed didn't exist before, it will be created for you.
To retrieve the widgets feed, make a HTTP GET request:

```bash
curl http://localhost:8000/feeds/widgets
```

## Development Environment

This project makes use of [Vagrant](http://vagrantup.com) and
[Docker](http://docker.com) for a consistent development environment.  The
configuration for these can be found in the `ops/` directory.

### Start docker-host

The first step is to start the `docker-host` vagrant machine, which is the
machine that will run the docker containers for this project.  You do this
from the `ops/` directory as follows:

```bash
cd ops
vagrant up docker-host
```

### Build the go.atompub binary

Now that the `docker-host` machine is ready to go, you can leverage the
defined `builder` container, which uses the
[golang:onbuild](https://hub.docker.com/_/golang/) container, to build
`go.atompub` in a consistent fashion.  To do that, run the following (from the
ops directory):

```bash
# From the ops/ directory
docker-atompub/build.sh
```

This script simply runs a `vagrant docker-run` command, which will run the
`ops/docker-atompub/container-build.sh` script from within the `builder`
container.  This build process will put the built go binary into your
repository's `bin/` directory, as `bin/go.atompub`.

Next, the `atompub` container will grab and run the built binary from that
directory.

## Run go.atompub container

Now you're ready t orun `go.atompub`.

```bash
# From ops/ directory

# go.atompub depends on postgres, so start that first
vagrant up postgres

# TODO: instructions on how to populate the database initially

# start the go.atompub container
vagrant up atompub
```

Now you should be able to hit `go.atompub` on localhost on port `8000`.
