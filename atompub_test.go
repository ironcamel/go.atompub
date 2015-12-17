package atompub

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/ironcamel/go.atompub"
)

func TestMain(m *testing.M) {
	rand.Seed(time.Now().UnixNano())
	startServer()
	os.Exit(m.Run())
}

func TestCreateFeed(t *testing.T) {
	feedTitle := fmt.Sprintf("test-feed-%d", rand.Intn(1000000000))
	url := _url("/feeds/" + feedTitle)

	res, err := http.Get(url)
	if err != nil || res.StatusCode != 404 {
		t.Error("expected 404 for", url, err, res)
		body, _ := ioutil.ReadAll(res.Body)
		t.Error(string(body))
	}

	entry := "<entry><title>foo</title><content>bar</content></entry>"
	buf := bytes.NewBufferString(entry)
	res, err = http.Post(url, "application/atom+xml", buf)
	if err != nil || res.StatusCode != 201 {
		t.Error("could not create feed", err, res)
	}

	res, err = http.Get(url)
	if err != nil || res.StatusCode != 200 {
		t.Error("could not get feed", err, res)
	}
	body, _ := ioutil.ReadAll(res.Body)
	fmt.Println(string(body))
}

func startServer() {
	go func() {
		dsn := "postgres://localhost/atompub?sslmode=disable"
		server := atompub.AtomPub{DSN: dsn}
		server.Start()
	}()
	for {
		res, err := http.Get("http://localhost:8000/status")
		if err == nil && res.StatusCode == 200 {
			break
		}
		fmt.Println("err:", err, "res:", res)
		fmt.Println("server not started yet, sleeping ...")
		time.Sleep(time.Millisecond * 10)
	}
}

func _url(uri string) string { return "http://localhost:8000" + uri }
