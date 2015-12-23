package atompub

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/ironcamel/go.atom"
	"github.com/stretchr/testify/require"
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
	require.Nil(t, err)
	require.Equal(t, res.StatusCode, 404)

	entry := "<entry><title>foo</title><content>bar</content></entry>"
	buf := bytes.NewBufferString(entry)
	res, err = http.Post(url, "application/atom+xml", buf)
	require.Nil(t, err)
	require.Equal(t, res.StatusCode, 201, "created feed")

	res, err = http.Get(url)
	require.Nil(t, err)
	require.Equal(t, res.StatusCode, 200, "got feed")

	feed, err := atom.DecodeFeed(res.Body)
	require.Nil(t, err, "parsed feed")
	require.Equal(t, len(feed.Entries), 1, "got 1 entry")
	require.Equal(t, feed.Title.Raw, feedTitle, "feed title")
}

func startServer() {
	go func() {
		dsn := "postgres://localhost/atompub?sslmode=disable"
		server := AtomPub{DSN: dsn}
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
