
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
