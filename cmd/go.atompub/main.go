package main

import (
	"os"

	"github.com/ironcamel/go.atompub"
)

func main() {
	server := atompub.AtomPub{
		Addr:    os.Getenv("GO_ATOMPUB_LISTEN_ADDR"),
		BaseURL: os.Getenv("GO_ATOMPUB_BASE_URL"),
		DSN:     os.Getenv("GO_ATOMPUB_DSN"),
	}
	server.Start()
}
