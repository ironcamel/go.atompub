package main

import (
	"os"

	"github.com/ironcamel/go.atompub"
)

func main() {
	server := atompub.AtomPub{
		BaseURL: os.Getenv("GO_ATOMPUB_BASE_URL"),
		DSN:     os.Getenv("GO_ATOMPUB_DSN"),
	}
	server.Start()
}
