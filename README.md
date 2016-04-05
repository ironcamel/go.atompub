
## Getting Started

Install go.atompub:

    go get github.com/ironcamel/go.atompub

Create the database:

    psql atompub < ./create-db.sql

Then run the server:

    go run ./cmd/go.atompub/main.go

## Usage

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

This will return an Atom feed:

```xml
<feed xmlns="http://www.w3.org/2005/Atom">
  <id>urn:uuid:12c428cf-49e2-4f90-9ac9-e6fffb25d73b</id>
  <title>widgets</title>
  <updated>2016-04-05T01:15:07Z</updated>
  <link href="http://localhost:8000/feeds/widgets?start-after=urn:uuid:13214e6b-3962-482b-bb1b-570790e4ff67" rel="next"/>
  <entry>
    <id>urn:uuid:13214e6b-3962-482b-bb1b-570790e4ff67</id>
    <title>allo</title>
    <updated>2016-04-05T01:15:07Z</updated>
    <content>{"foo":"bar"}</content>
  </entry>
</feed>
```

Clients can request entries that came after the last entry they processed by
providing the id of the last entry as the start-after query parameter:

```bash
curl http://localhost/atombus/feeds/widgets?start-after=42
```

## API Documentation

See: https://godoc.org/github.com/ironcamel/go.atompub
