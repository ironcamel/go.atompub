
## Getting Started

Install go.atompub:

    go get github.com/ironcamel/go.atompub

Create the database:

    psql < ./create-db.sql

Then run the server:

    GO_ATOMPUB_DSN='postgres://...' go run ./cmd/go.atompub/main.go

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
You can retrieve the widgets feed like so:

```bash
curl http://localhost:8000/feeds/widgets
```

Clients can request only entries that came after the last entry they processed.
They can do this by providing the id of the last message as the start-after
query parameter:

    $ curl http://localhost/atombus/feeds/widgets?start-after=42

