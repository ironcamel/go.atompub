package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/ironcamel/go.atompub"
)

func main() {
	server := atompub.AtomPub{}

	var port int
	envPort := os.Getenv("GO_ATOMPUB_PORT")
	if envPort == "" {
		envPort = os.Getenv("PORT")
	}
	if envPort == "" {
		port = 8000
	} else {
		var err error
		if port, err = strconv.Atoi(envPort); err != nil {
			log.Fatal("Invalid port value: ", envPort)
		}
	}
	server.Port = port

	server.BaseURL = os.Getenv("GO_ATOMPUB_BASE_URL")
	if server.BaseURL == "" {
		server.BaseURL = fmt.Sprint("http://localhost:", port)
	}

	server.DSN = os.Getenv("GO_ATOMPUB_DSN")

	server.Start()
}
